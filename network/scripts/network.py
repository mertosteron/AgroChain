#!/usr/bin/env python3
"""Small standard-library checks; no Docker SDK or YAML dependency required."""
import base64
import hashlib
import json
import os
import re
from pathlib import Path
import shutil
import socket
import ssl
import string
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
ORGS = ("producer", "logistics", "retailer", "regulator")
SERVICES = ("orderer1", "orderer2", "orderer3", *ORGS)
def ledger_prefix():
    value = os.environ.get("AGROCHAIN_LEDGER_VOLUME_PREFIX", "agrochain")
    if not re.fullmatch(r"agrochain(?:-stage7-[0-9a-f]{12})?", value):
        raise ValueError("Invalid isolated ledger volume prefix")
    return value


VOLUMES = tuple(f"{ledger_prefix()}-{s}-data" for s in SERVICES)
GENERATED = (
    "organizations/peerOrganizations", "organizations/ordererOrganizations",
    "channel-artifacts/agrochannel.block", "channel-artifacts/generated-manifest.json",
    "connection-profiles/generated", "runtime",
)


def require(condition, message):
    if not condition:
        raise ValueError(message)


def run(*args):
    return subprocess.check_output(args, text=True).strip()


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def source_digest(root=ROOT):
    inputs = [root / ".env.example", root / "config/configtx.yaml",
              root / "config/crypto-config.yaml", root / "connection-profiles/organization.json.template"]
    return hashlib.sha256(b"".join(p.read_bytes() for p in inputs)).hexdigest()


def safe_targets(root=ROOT):
    """Never follow symlinks (including parent directories) during cleanup."""
    root = root.absolute()
    require(root == root.resolve(), "Refusing symlinked network root")
    paths = [root / name for name in GENERATED]
    for path in paths:
        for parent in (path, *path.parents):
            if parent == root:
                break
            require(not parent.is_symlink(), f"Refusing symlink: {parent}")
        if path.is_dir():
            for entry in path.rglob("*"):
                require(not entry.is_symlink(), f"Refusing generated symlink: {entry}")
                require(entry.is_file() or entry.is_dir(), f"Unexpected generated object: {entry}")
    return paths


def generated_files(root=ROOT):
    safe_targets(root)
    result = []
    for name in GENERATED[:3] + (GENERATED[4],):
        path = root / name
        result.extend(sorted(path.rglob("*")) if path.is_dir() else [path])
    return {str(p.relative_to(root)): digest(p) for p in result if p.is_file()}


def record_generated():
    data = {"sourceSha256": source_digest(), "files": generated_files(), "ledgerVolumePrefix": ledger_prefix()}
    (ROOT / GENERATED[3]).write_text(json.dumps(data, indent=2) + "\n")


def validate_generated():
    safe_targets()
    manifest = ROOT / GENERATED[3]
    require(manifest.is_file(), "Run make generate first (or clean-generated after interrupted generation)")
    data = json.loads(manifest.read_text())
    require(data.get("ledgerVolumePrefix", "agrochain") == ledger_prefix(), "Ledger volume prefix differs from generated identities")
    require(data["sourceSha256"] == source_digest(), "Generation inputs changed; explicit stop/clean/bootstrap required")
    require(data["files"] == generated_files(), "Generated artifacts changed or missing; restore or explicitly stop/clean/bootstrap")
    require((ROOT / GENERATED[2]).stat().st_size > 0, "Missing channel genesis block")
    for org in ORGS:
        profile_path = ROOT / "connection-profiles/generated" / f"{org}.json"
        profile = json.loads(profile_path.read_text())
        require(profile["organizations"][f"{org.title()}MSP"]["mspid"] == f"{org.title()}MSP", "Profile MSP mismatch")
        for endpoint in (*profile["peers"].values(), *profile["orderers"].values()):
            require(endpoint["url"].startswith("grpcs://"), "Profile TLS disabled")
            ca = Path(endpoint["tlsCACerts"]["path"])
            require(not ca.is_absolute(), "Profile contains absolute certificate path")
            require((profile_path.parent / ca).is_file(), "Missing profile CA")


def profiles():
    template = string.Template((ROOT / "connection-profiles/organization.json.template").read_text())
    directory = ROOT / "connection-profiles/generated"
    directory.mkdir(exist_ok=True)
    for i, org in enumerate(ORGS):
        rendered = template.substitute(ORG=org, MSP=f"{org.title()}MSP", PEER_PORT=7051 + 1000 * i)
        data = json.loads(rendered)
        (directory / f"{org}.json").write_text(json.dumps(data, indent=2) + "\n")


def check_docker():
    server = run("docker", "version", "--format", "{{.Server.Version}}")
    compose = run("docker", "compose", "version", "--short").lstrip("v")
    require(int(server.split(".")[0]) >= 24, "Docker Engine >=24 required")
    parts = tuple(int(p) for p in compose.split("-")[0].split(".")[:2])
    require(parts >= (2, 20), "Compose >=2.20 required (newer plugin majors accepted)")
    print(f"Docker Engine {server}; Compose {compose}")


def inspect_optional(kind, name):
    result = subprocess.run(["docker", kind, "inspect", name], capture_output=True, text=True)
    if result.returncode:
        # Daemon availability was checked; distinguish absence from permission failure.
        absent = "no such" in result.stderr.lower() or f"network {name} not found" in result.stderr
        require(absent, result.stderr.strip())
        return None
    return json.loads(result.stdout)[0]


def probe_port(port):
    with socket.socket() as sock:
        # Recently closed TLS connections can leave TIME_WAIT after docker down.
        # Reuse permits that state, but still refuses an active TCP listener.
        sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        try:
            sock.bind(("127.0.0.1", port))
        except OSError as exc:
            raise ValueError(f"Port {port} unavailable: {exc}") from exc


def check_ports():
    for volume in VOLUMES:
        info = inspect_optional("volume", volume)
        if info:
            require((info.get("Labels") or {}).get("com.docker.compose.project") == "agrochain", f"Foreign volume name collision: {volume}")
            require((ROOT / GENERATED[3]).is_file(), "Ledger volumes exist without generated identities; restore identities or stop/clean-generated")
    network_info = inspect_optional("network", "agrochain-fabric")
    if network_info:
        require((network_info.get("Labels") or {}).get("com.docker.compose.project") == "agrochain", "Foreign Docker network name collision")
    expected = [(s, 7051 + i * 1000) for i, s in enumerate(ORGS)]
    expected += [(f"orderer{i}", base + i * 1000) for i in range(1, 4) for base in (6050, 6053)]
    for service, port in expected:
        info = inspect_optional("container", f"agrochain-{service}")
        if info:
            require(info["Config"]["Labels"].get("com.docker.compose.project") == "agrochain", "Container name owned by another project")
        if info and info["State"]["Running"]:
            bindings = info["NetworkSettings"]["Ports"]
            require(any(b["HostIp"] == "127.0.0.1" and b["HostPort"] == str(port)
                        for values in bindings.values() for b in (values or [])), f"Unexpected port binding: {service}")
            continue
        probe_port(port)


def verify_nodes():
    version = os.environ["FABRIC_VERSION"]
    ids = run("docker", "ps", "-aq", "--filter", "label=com.docker.compose.project=agrochain").splitlines()
    require(len(ids) == 7, "Expected exactly seven AgroChain node containers")
    for service in SERVICES:
        node = inspect_optional("container", f"agrochain-{service}")
        require(node is not None, f"Missing {service}")
        kind = "orderer" if service.startswith("orderer") else "peer"
        require(node["Config"]["Image"] == f"hyperledger/fabric-{kind}:{version}", f"Wrong image: {service}")
        require(node["Config"]["Labels"]["com.docker.compose.project"] == "agrochain", "Wrong project")
        require(node["State"]["Running"] and node["State"]["Health"]["Status"] == "healthy", f"Unhealthy {service}")
        env = dict(v.split("=", 1) for v in node["Config"]["Env"] if "=" in v)
        prefix = "ORDERER_GENERAL" if kind == "orderer" else "CORE_PEER"
        expected_msp = "OrdererMSP" if kind == "orderer" else f"{service.title()}MSP"
        require(env[f"{prefix}_TLS_ENABLED"] == "true", f"TLS disabled: {service}")
        require(env[f"{prefix}_LOCALMSPID"] == expected_msp, f"Wrong node MSP: {service}")
        if kind == "peer":
            require(env["CORE_PEER_GATEWAY_ENABLED"] == "true", f"Gateway disabled: {service}")
    print("PASS seven pinned, healthy, TLS-enabled nodes and local MSP IDs")


def tls_connect(port, hostname, ca, client=None):
    context = ssl.create_default_context(cafile=str(ca))
    if client:
        context.load_cert_chain(str(client / "client.crt"), str(client / "client.key"))
    with socket.create_connection(("127.0.0.1", port), timeout=5) as raw:
        with context.wrap_socket(raw, server_hostname=hostname) as secure:
            require(secure.version() in ("TLSv1.2", "TLSv1.3"), "Unsupported TLS version")
            if client:
                secure.sendall(b"GET /participation/v1/channels HTTP/1.0\r\nHost: localhost\r\n\r\n")
                require(b" 200 " in secure.recv(4096).split(b"\r\n")[0], "Orderer mTLS admin HTTP failure")


def verify_tls():
    for i, org in enumerate(ORGS):
        host = f"peer0.{org}.agrochain.test"
        ca = ROOT / f"organizations/peerOrganizations/{org}.agrochain.test/peers/{host}/tls/ca.crt"
        tls_connect(7051 + i * 1000, host, ca)
    base = ROOT / "organizations/ordererOrganizations/orderer.agrochain.test"
    for i in range(1, 4):
        host = f"orderer{i}.orderer.agrochain.test"
        ca = base / f"orderers/{host}/tls/ca.crt"
        tls_connect(6050 + 1000 * i, host, ca)
        tls_connect(6053 + 1000 * i, host, ca, base / "users/Admin@orderer.agrochain.test/tls")
        # Handshake may complete under TLS 1.3 before the missing-client-cert alert.
        context = ssl.create_default_context(cafile=str(ca))
        try:
            with socket.create_connection(("127.0.0.1", 6053 + 1000 * i), timeout=5) as raw:
                with context.wrap_socket(raw, server_hostname=host) as secure:
                    secure.sendall(b"GET /participation/v1/channels HTTP/1.0\r\nHost: localhost\r\n\r\n")
                    response = secure.recv(4096)
                    require(not response, "Admin endpoint accepted a client without a TLS certificate")
        except ssl.SSLError:
            pass
    wrong_ca = ROOT / "organizations/peerOrganizations/logistics.agrochain.test/peers/peer0.logistics.agrochain.test/tls/ca.crt"
    try:
        tls_connect(7051, "peer0.producer.agrochain.test", wrong_ca)
    except ssl.SSLCertVerificationError:
        pass
    else:
        raise ValueError("Peer accepted incorrect TLS trust root")
    print("PASS four peer TLS, three orderer TLS, three mutual-TLS admin endpoints; wrong CA and unauthenticated admin rejected")


def verify_config(data, root=ROOT):
    envelope = data["data"]["data"][0]
    require(envelope["payload"]["header"]["channel_header"]["channel_id"] == "agrochannel", "Wrong channel ID")
    group = envelope["payload"]["data"]["config"]["channel_group"]
    app = group["groups"]["Application"]["groups"]
    require(set(app) == {f"{org.title()}MSP" for org in ORGS}, "Wrong application MSP set")
    for org in ORGS:
        msp = f"{org.title()}MSP"
        values = app[msp]["values"]
        conf = values["MSP"]["value"]["config"]
        require(conf["name"] == msp and conf["fabric_node_ous"]["enable"], f"Invalid MSP/NodeOUs: {msp}")
        cert = root / f"organizations/peerOrganizations/{org}.agrochain.test/msp/cacerts/ca.{org}.agrochain.test-cert.pem"
        require([base64.b64decode(c) for c in conf["root_certs"]] == [cert.read_bytes()], f"MSP CA mismatch: {msp}")
        require(values["AnchorPeers"]["value"]["anchor_peers"] == [{"host": f"peer0.{org}.agrochain.test", "port": 7051}], f"Wrong anchor: {msp}")
    orderer = group["groups"]["Orderer"]
    require(set(orderer["groups"]) == {"OrdererMSP"}, "Wrong ordering governance")
    consensus = orderer["values"]["ConsensusType"]["value"]
    require(consensus["type"] == "etcdraft", "Expected Raft consensus")
    metadata = consensus["metadata"]
    if isinstance(metadata, str):
        metadata = json.loads(subprocess.check_output(["configtxlator", "proto_decode", "--type", "etcdraft.ConfigMetadata"], input=base64.b64decode(metadata)))
    consenters = metadata["consenters"]
    require(len(consenters) == 3, "Expected three Raft consenters")
    require({c["host"] for c in consenters} == {f"orderer{i}.orderer.agrochain.test" for i in range(1, 4)}, "Wrong consenter hosts")
    for consenter in consenters:
        require(consenter["port"] == 7050, "Wrong Raft port")
        cert = root / f"organizations/ordererOrganizations/orderer.agrochain.test/orderers/{consenter['host']}/tls/server.crt"
        require(all(base64.b64decode(consenter[k]) == cert.read_bytes() for k in ("client_tls_cert", "server_tls_cert")), "Raft TLS certificate mismatch")
    print("PASS channel configuration: four MSP roots/NodeOUs, four anchors, OrdererMSP, three Raft consenters/TLS certificates")


def cleanup_preflight():
    safe_targets()
    manifest = ROOT / GENERATED[3]
    if manifest.exists():
        require(json.loads(manifest.read_text()).get("ledgerVolumePrefix", "agrochain") == ledger_prefix(), "Ledger volume prefix differs from generated identities")
    require(not run("docker", "ps", "-aq", "--filter", "label=com.docker.compose.project=agrochain"), "Run make network-down before clean-generated")
    for service in SERVICES:
        require(inspect_optional("container", f"agrochain-{service}") is None, "Named container still exists; cleanup refused")
    for name in VOLUMES:
        info = inspect_optional("volume", name)
        if info:
            require((info.get("Labels") or {}).get("com.docker.compose.project") == "agrochain", f"Foreign volume: {name}")
            require(not run("docker", "ps", "-aq", "--filter", f"volume={name}"), f"Volume in use: {name}")


def clean_generated():
    cleanup_preflight()
    for name in VOLUMES:
        if inspect_optional("volume", name):
            subprocess.run(["docker", "volume", "rm", name], check=True)
    for path in safe_targets():
        if path.is_dir():
            # Go module caches contain owner-read-only directories. These are
            # already validated generated targets; allow unlinking their children.
            for directory in (path, *path.rglob("*")):
                if directory.is_dir():
                    directory.chmod(directory.stat().st_mode | 0o700)
            shutil.rmtree(path)
        elif path.exists():
            path.unlink()


def main():
    command = sys.argv[1]
    if command == "require-empty":
        require(not any(p.exists() for p in safe_targets() if p.name != "runtime"), "Partial generated artifacts exist; use explicit stop/clean-generated before regeneration")
    elif command == "verify-config":
        verify_config(json.loads(Path(sys.argv[2]).read_text()))
    else:
        commands = {"profiles": profiles, "record-generated": record_generated,
                    "validate-generated": validate_generated, "check-docker": check_docker,
                    "check-ports": check_ports, "verify-nodes": verify_nodes, "verify-tls": verify_tls,
                    "cleanup-preflight": cleanup_preflight, "clean-generated": clean_generated}
        commands[command]()


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, KeyError, subprocess.CalledProcessError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        sys.exit(1)
