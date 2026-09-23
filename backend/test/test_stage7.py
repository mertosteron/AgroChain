import unittest
from stage7_measure import summarize


class StatisticsTests(unittest.TestCase):
    def test_nearest_rank_and_even_median(self):
        s = summarize(list(range(1, 31)))
        self.assertEqual((s["n"], s["medianMs"], s["p95Ms"]), (30, 15.5, 29))

    def test_invalid_and_small_samples(self):
        self.assertEqual(summarize([7])["p95Ms"], 7)
        for values in ([], [-1], [float("nan")], [float("inf")]):
            with self.assertRaises(ValueError):
                summarize(values)


if __name__ == "__main__":
    unittest.main()
