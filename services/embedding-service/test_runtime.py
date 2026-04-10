import pathlib
import sys
import unittest
from unittest.mock import patch

import numpy as np

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent))

import embedding_runtime


class FakeModel:
    def embed(self, _texts):
        yield np.array([0.5, -0.2, 0.1], dtype=np.float32)


class EmbeddingRuntimeTests(unittest.TestCase):
    def test_embed_text_returns_padded_embedding(self):
        with patch.object(embedding_runtime, "get_model", return_value=FakeModel()):
            result = embedding_runtime.embed_text("Café Marley 170gr")

        self.assertEqual(result.base_dimensions, 3)
        self.assertEqual(result.dimensions, embedding_runtime.TARGET_DIMENSIONS)
        self.assertAlmostEqual(result.embedding[0], 0.5, places=6)
        self.assertAlmostEqual(result.embedding[1], -0.2, places=6)
        self.assertAlmostEqual(result.embedding[2], 0.1, places=6)

    def test_embed_text_rejects_empty_normalized_text(self):
        with self.assertRaises(ValueError):
            embedding_runtime.embed_text("!!!")


if __name__ == "__main__":
    unittest.main()
