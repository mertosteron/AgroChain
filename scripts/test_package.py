import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location("package_demo", Path(__file__).with_name("package-demo.py"))
package = importlib.util.module_from_spec(spec)
spec.loader.exec_module(package)


class PackageTests(unittest.TestCase):
    def test_pem_block_is_rejected_but_parser_literal_is_not(self):
        self.assertTrue(package.contains_private_key(b"-----BEGIN PRIVATE KEY-----\nsynthetic-test\n"))
        self.assertTrue(package.contains_private_key(b"  -----BEGIN EC PRIVATE KEY-----\nsynthetic-test\n"))
        self.assertFalse(package.contains_private_key(b'if b"-----BEGIN PRIVATE KEY-----" in data:'))

    def test_credentials_and_runtime_are_excluded(self):
        for name in ["network/runtime/source-keys/CKS.pem", "network/.env", "backend/tokens.json", "backend/key_sk", "backend/private.key", "network/organizations/admin/cert.pem", "backend/target/server.jar"]:
            self.assertFalse(package.permitted(name), name)

    def test_source_and_placeholder_survive(self):
        for name in ["network/.env.example", "network/organizations/.gitkeep", "scripts/demo.py", "docs/competition/AgroChain.pptx", "backend/src/main/java/org/agrochain/Api.java"]:
            self.assertTrue(package.permitted(name), name)

    def test_path_escape_is_rejected(self):
        for name in ["/tmp/secret", "docs/../../secret", ".git/config", "dist/archive.tar.gz"]:
            self.assertFalse(package.permitted(name), name)


if __name__ == "__main__":
    unittest.main()
