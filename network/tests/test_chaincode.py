import gzip
import importlib.util
import io
import json
from pathlib import Path
import tarfile
import unittest

ROOT = Path(__file__).resolve().parents[2]
spec = importlib.util.spec_from_file_location("chaincode_package", ROOT / "network/scripts/chaincode-package.py")
packaging = importlib.util.module_from_spec(spec)
spec.loader.exec_module(packaging)


class PackagingTests(unittest.TestCase):
    def test_archives_reproducible_and_sorted(self):
        a = packaging.archive({"z": b"last", "a": b"first"})
        b = packaging.archive({"a": b"first", "z": b"last"})
        self.assertEqual(a, b)
        with tarfile.open(fileobj=io.BytesIO(a), mode="r:gz") as archive:
            self.assertEqual(archive.getnames(), ["a", "z"])
            for member in archive.getmembers():
                self.assertEqual((member.mtime, member.uid, member.gid, member.mode), (0, 0, 0, 0o644))

    def test_release_pins_and_no_docker_socket(self):
        compose = (ROOT / "network/compose/compose-chaincode.yaml").read_text()
        self.assertNotIn("docker.sock", compose)
        self.assertIn("read_only: true", compose)
        self.assertIn("cap_drop: [ALL]", compose)
        dockerfile = (ROOT / "chaincode/agrochain/Dockerfile").read_text()
        self.assertIn("FROM scratch", dockerfile)
        self.assertNotIn("latest", dockerfile)

    def test_production_dispatcher_has_no_test_provider_or_runtime_enable_flag(self):
        directory = ROOT / "chaincode/agrochain/internal/agrochain"
        for path in directory.glob("*.go"):
            if path.name.endswith("_test.go"):
                continue
            source = path.read_text()
            for prohibited in ("fixtureEvidence", '"net/http"', '"os"', '"math/rand"', '"crypto/rand"', "time.Now(", "time.Sleep("):
                self.assertNotIn(prohibited, source, str(path))
        contract = (directory / "contract.go").read_text()
        self.assertIn("realEvidence{", contract)
        self.assertIn("getConfig(s)", contract)
        self.assertNotIn("closedEvidence{}", contract)

    def test_existing_packages_never_contain_private_keys(self):
        # Package byte safety is exercised when generated packages exist; absence
        # on a clean checkout is not represented as a live deployment assertion.
        for path in (ROOT / "network/runtime/chaincode").glob("*/chaincode.tar.gz"):
            with tarfile.open(path, "r:gz") as outer:
                self.assertEqual(set(outer.getnames()), {"metadata.json", "code.tar.gz"})
                with tarfile.open(fileobj=io.BytesIO(outer.extractfile("code.tar.gz").read()), mode="r:gz") as inner:
                    self.assertEqual(inner.getnames(), ["connection.json"])
                    raw = inner.extractfile("connection.json").read()
                    self.assertNotIn(b"PRIVATE KEY", raw)
                    connection = json.loads(raw)
                    self.assertTrue(connection["tls_required"])
                    self.assertIn("BEGIN CERTIFICATE", connection["root_cert"])
