import base64
import importlib.util
import json
import os
import socket
from pathlib import Path
import string
import subprocess
import tempfile
import unittest
from unittest.mock import patch

NETWORK = Path(__file__).resolve().parents[1]
spec = importlib.util.spec_from_file_location("network_helpers", NETWORK / "scripts/network.py")
network = importlib.util.module_from_spec(spec)
spec.loader.exec_module(network)


class SafetyTests(unittest.TestCase):
    def test_isolated_volume_prefix_is_strict(self):
        for value in ("", "other", "agrochain-stage7-../", "agrochain-stage7-123"):
            with patch.dict(os.environ, AGROCHAIN_LEDGER_VOLUME_PREFIX=value):
                with self.assertRaises(ValueError):
                    network.ledger_prefix()
        with patch.dict(os.environ, AGROCHAIN_LEDGER_VOLUME_PREFIX="agrochain-stage7-012345abcdef"):
            self.assertEqual(network.ledger_prefix(), "agrochain-stage7-012345abcdef")

    def test_cleanup_rejects_identity_volume_mismatch_before_docker(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            manifest = root / network.GENERATED[3]
            manifest.parent.mkdir(parents=True)
            manifest.write_text(json.dumps({"ledgerVolumePrefix": "agrochain-stage7-012345abcdef"}))
            with patch.object(network, "ROOT", root), patch.object(network, "run") as docker:
                with self.assertRaisesRegex(ValueError, "prefix differs"):
                    network.cleanup_preflight()
                docker.assert_not_called()

    def test_port_probe_rejects_live_listener(self):
        with socket.socket() as listener:
            listener.bind(("127.0.0.1", 0))
            listener.listen()
            with self.assertRaisesRegex(ValueError, "unavailable"):
                network.probe_port(listener.getsockname()[1])

    def test_port_probe_accepts_recently_closed_connection(self):
        with socket.socket() as listener:
            listener.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            listener.bind(("127.0.0.1", 0))
            port = listener.getsockname()[1]
            listener.listen()
            with socket.create_connection(("127.0.0.1", port), timeout=2) as client:
                server, _ = listener.accept()
                server.close()
                self.assertEqual(client.recv(1), b"")
        network.probe_port(port)

    def test_missing_docker_network_is_normal_on_clean_bootstrap(self):
        result = subprocess.CompletedProcess([], 1, "[]", "Error response from daemon: network agrochain-fabric not found")
        with patch.object(network.subprocess, "run", return_value=result):
            self.assertIsNone(network.inspect_optional("network", "agrochain-fabric"))

    def test_docker_permission_failure_is_not_treated_as_absence(self):
        result = subprocess.CompletedProcess([], 1, "", "permission denied accessing docker.sock")
        with patch.object(network.subprocess, "run", return_value=result):
            with self.assertRaisesRegex(ValueError, "permission denied"):
                network.inspect_optional("network", "agrochain-fabric")

    def test_cleanup_refuses_target_symlink(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            (root / "organizations").mkdir()
            (root / "organizations/peerOrganizations").symlink_to(root / "valuable")
            with self.assertRaisesRegex(ValueError, "symlink"):
                network.safe_targets(root)

    def test_cleanup_refuses_parent_symlink(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            (root / "organizations").symlink_to(root / "valuable")
            with self.assertRaisesRegex(ValueError, "symlink"):
                network.safe_targets(root)

    def test_cleanup_refuses_nested_symlink(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            target = root / "organizations/peerOrganizations"
            target.mkdir(parents=True)
            (target / "key").symlink_to(root / "valuable")
            with self.assertRaisesRegex(ValueError, "symlink"):
                network.safe_targets(root)

    def test_cleanup_preserves_source_tools_and_unknown_files(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            for name in ("organizations/.gitkeep", "organizations/user-notes", "config/configtx.yaml", "tools/bin/peer", "runtime/check.log"):
                p = root / name
                p.parent.mkdir(parents=True, exist_ok=True)
                p.write_text("keep unless generated")
            targets = network.safe_targets(root)
            with patch.object(network, "cleanup_preflight"), patch.object(network, "inspect_optional", return_value=None), patch.object(network, "safe_targets", return_value=targets):
                network.clean_generated()
            self.assertFalse((root / "runtime").exists())
            for name in ("organizations/.gitkeep", "organizations/user-notes", "config/configtx.yaml", "tools/bin/peer"):
                self.assertTrue((root / name).is_file())

    def test_cleanup_removes_readonly_generated_go_cache(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            cache = root / "runtime/go-mod/example@v1"
            cache.mkdir(parents=True)
            (cache / "go.mod").write_text("module example")
            cache.chmod(0o555)
            targets = network.safe_targets(root)
            with patch.object(network, "cleanup_preflight"), patch.object(network, "inspect_optional", return_value=None), patch.object(network, "safe_targets", return_value=targets):
                network.clean_generated()
            self.assertFalse((root / "runtime").exists())

    def test_cleanup_rejects_foreign_volume(self):
        def inspect(kind, name):
            return {"Labels": {"com.docker.compose.project": "someone-else"}} if kind == "volume" else None
        with patch.object(network, "run", return_value=""), patch.object(network, "inspect_optional", side_effect=inspect):
            with self.assertRaisesRegex(ValueError, "Foreign volume"):
                network.cleanup_preflight()

    def test_cleanup_rejects_existing_containers(self):
        with patch.object(network, "run", return_value="container-id"):
            with self.assertRaisesRegex(ValueError, "network-down"):
                network.cleanup_preflight()


class ChannelChecks(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.app = {}
        for org in network.ORGS:
            cert = self.root / f"organizations/peerOrganizations/{org}.agrochain.test/msp/cacerts/ca.{org}.agrochain.test-cert.pem"
            cert.parent.mkdir(parents=True)
            cert.write_bytes(f"test CA {org}".encode())
            self.app[f"{org.title()}MSP"] = {"values": {
                "MSP": {"value": {"config": {"name": f"{org.title()}MSP", "fabric_node_ous": {"enable": True}, "root_certs": [base64.b64encode(cert.read_bytes()).decode()]}}},
                "AnchorPeers": {"value": {"anchor_peers": [{"host": f"peer0.{org}.agrochain.test", "port": 7051}]}}}}
        self.consenters = []
        for i in range(1, 4):
            host = f"orderer{i}.orderer.agrochain.test"
            cert = self.root / f"organizations/ordererOrganizations/orderer.agrochain.test/orderers/{host}/tls/server.crt"
            cert.parent.mkdir(parents=True)
            cert.write_bytes(f"test TLS {i}".encode())
            encoded = base64.b64encode(cert.read_bytes()).decode()
            self.consenters.append({"host": host, "port": 7050, "client_tls_cert": encoded, "server_tls_cert": encoded})
        self.data = {"data": {"data": [{"payload": {"header": {"channel_header": {"channel_id": "agrochannel"}}, "data": {"config": {"channel_group": {"groups": {
            "Application": {"groups": self.app}, "Orderer": {"groups": {"OrdererMSP": {}}, "values": {"ConsensusType": {"value": {"type": "etcdraft", "metadata": {"consenters": self.consenters}}}}}
        }}}}}}]}}

    def test_valid_configuration(self):
        network.verify_config(self.data, self.root)

    def test_missing_msp_rejected(self):
        del self.app["RegulatorMSP"]
        with self.assertRaisesRegex(ValueError, "MSP set"):
            network.verify_config(self.data, self.root)

    def test_wrong_anchor_rejected(self):
        self.app["ProducerMSP"]["values"]["AnchorPeers"]["value"]["anchor_peers"][0]["port"] = 99
        with self.assertRaisesRegex(ValueError, "Wrong anchor"):
            network.verify_config(self.data, self.root)

    def test_missing_consenter_rejected(self):
        self.consenters.pop()
        with self.assertRaisesRegex(ValueError, "three Raft"):
            network.verify_config(self.data, self.root)

    def test_wrong_consenter_certificate_rejected(self):
        self.consenters[0]["server_tls_cert"] = base64.b64encode(b"wrong cert").decode()
        with self.assertRaisesRegex(ValueError, "certificate mismatch"):
            network.verify_config(self.data, self.root)

    def test_wrong_msp_root_rejected(self):
        self.app["ProducerMSP"]["values"]["MSP"]["value"]["config"]["root_certs"] = []
        with self.assertRaisesRegex(ValueError, "CA mismatch"):
            network.verify_config(self.data, self.root)


class SourceChecks(unittest.TestCase):
    def test_shell_syntax_and_strict_mode(self):
        for path in (NETWORK / "scripts").glob("*.sh"):
            self.assertIn("set -Eeuo pipefail", path.read_text())
            subprocess.run(["bash", "-n", str(path)], check=True)

    def test_compose_topology_tls_and_persistence(self):
        raw = subprocess.check_output(["docker", "compose", "--env-file", str(NETWORK / ".env.example"), "-f", str(NETWORK / "compose/compose-agrochain.yaml"), "config", "--format", "json"], text=True)
        config = json.loads(raw)
        self.assertEqual(set(config["services"]), set(network.SERVICES))
        host_ports = []
        for name, service in config["services"].items():
            self.assertNotIn(":latest", service["image"])
            self.assertIn("healthcheck", service)
            self.assertTrue(any(v["type"] == "volume" for v in service["volumes"]))
            self.assertFalse(any("docker.sock" in v["source"] for v in service["volumes"]))
            for port in service["ports"]:
                self.assertEqual(port["host_ip"], "127.0.0.1")
                host_ports.append(port["published"])
            key = "ORDERER_GENERAL_TLS_ENABLED" if name.startswith("orderer") else "CORE_PEER_TLS_ENABLED"
            self.assertEqual(service["environment"][key], "true")
        self.assertEqual(len(host_ports), len(set(host_ports)))

    def test_collection_membership_and_endorsers(self):
        collections = json.loads((NETWORK / "config/collections.json").read_text())
        self.assertEqual([c["name"] for c in collections], ["tradePrivate", "freightPrivate", "retailAuditPrivate"])
        self.assertNotIn("LogisticsMSP", collections[0]["policy"])
        self.assertNotIn("ProducerMSP", collections[1]["policy"])
        self.assertNotIn("ProducerMSP", collections[2]["policy"])
        self.assertNotIn("LogisticsMSP", collections[2]["policy"])
        for c in collections:
            self.assertTrue(c["memberOnlyRead"] and c["memberOnlyWrite"])
            self.assertEqual(c["endorsementPolicy"]["signaturePolicy"], "AND('RetailerMSP.peer','RegulatorMSP.peer')")

    def test_connection_profile_template_all_orgs(self):
        template = string.Template((NETWORK / "connection-profiles/organization.json.template").read_text())
        for i, org in enumerate(network.ORGS):
            text = template.substitute(ORG=org, MSP=f"{org.title()}MSP", PEER_PORT=7051 + i * 1000)
            self.assertNotIn("PRIVATE KEY", text)
            profile = json.loads(text)
            self.assertEqual(len(profile["orderers"]), 3)
            for item in (*profile["peers"].values(), *profile["orderers"].values()):
                self.assertTrue(item["url"].startswith("grpcs://localhost:"))
                self.assertFalse(Path(item["tlsCACerts"]["path"]).is_absolute())


if __name__ == "__main__":
    unittest.main()
