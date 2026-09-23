import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch
from demo import Client, fixture, verify


class DemoTests(unittest.TestCase):
    def test_fixture_is_repeatable_and_distinct(self):
        operations = set()
        for scenario, price in [("NORMAL", 2800), ("SUSPICIOUS", 3700)]:
            batch, lot, commands = fixture(scenario, price)
            self.assertEqual((batch, lot, commands), fixture(scenario, price))
            self.assertEqual(len(commands), 7)
            self.assertEqual(commands[-1][2]["privateInput"]["offeredPriceKurusPerKg"], price)
            for i, (_, _, req) in enumerate(commands):
                self.assertEqual(req["command"]["expectedVersion"], i)
                operations.add(req["command"]["operationId"])
        self.assertEqual(len(operations), 14)

    def test_remote_origin_rejected_before_credentials_read(self):
        for origin in ["http://example.org", "https://localhost", "http://localhost@evil.test", "http://localhost/path", "http://localhost?x=1"]:
            with self.assertRaises(ValueError):
                Client(origin, "/nonexistent/token/file")

    def test_pending_reuses_exact_operation(self):
        client = object.__new__(Client)
        request = fixture("NORMAL", 2800)[2][0][2]
        with patch.object(client, "call", side_effect=[(202, {"status": "SUBMITTED_UNKNOWN"}), (200, {"status": "COMMITTED", "txId": "tx"})]) as call, patch("demo.time.sleep"):
            self.assertEqual(client.submit("producer", "/batches", request), "tx")
            self.assertEqual(call.call_args_list[0], call.call_args_list[1])

    def test_auth_failure_is_not_retried(self):
        client = object.__new__(Client)
        with patch.object(client, "call", return_value=(403, {"code": "UNAUTHORIZED_ROLE"})) as call:
            with self.assertRaisesRegex(RuntimeError, "UNAUTHORIZED_ROLE"):
                client.submit("producer", "/batches", fixture("NORMAL", 2800)[2][0][2])
            self.assertEqual(call.call_count, 1)


if __name__ == "__main__":
    unittest.main()
