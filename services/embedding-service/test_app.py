import pathlib
import sys
import unittest

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

from text_normalization import canonicalize_token, normalize_text, to_passage_text


class EmbeddingServiceNormalizationTests(unittest.TestCase):
    def test_normalize_text_applies_local_business_aliases(self):
        self.assertEqual(
            normalize_text("Café Marley instant coffee 170gr"),
            "cafe marley instantaneo cafe 170 g",
        )

    def test_canonicalize_token_maps_common_variants(self):
        self.assertEqual(canonicalize_token("coffee"), "cafe")
        self.assertEqual(canonicalize_token("coffe"), "cafe")
        self.assertEqual(canonicalize_token("grs"), "g")

    def test_to_passage_text_prefixes_normalized_value(self):
        self.assertEqual(
            to_passage_text("Café Nestle 170gr"),
            "passage: cafe nestle 170 g",
        )


if __name__ == "__main__":
    unittest.main()
