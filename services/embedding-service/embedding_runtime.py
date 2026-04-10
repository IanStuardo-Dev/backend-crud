import os
from dataclasses import dataclass
from functools import lru_cache
from typing import Any, List

import numpy as np

from text_normalization import normalize_text, to_passage_text


MODEL_NAME = os.getenv("EMBEDDING_MODEL", "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2")
TARGET_DIMENSIONS = int(os.getenv("EMBEDDING_TARGET_DIMENSIONS", "1536"))
DEVICE = os.getenv("EMBEDDING_DEVICE", "cpu")
GRPC_PORT = int(os.getenv("GRPC_PORT", "50051"))


@dataclass
class EmbeddingResult:
    embedding: List[float]
    dimensions: int
    base_dimensions: int
    model: str


@lru_cache
def get_model() -> Any:
    from fastembed import TextEmbedding

    return TextEmbedding(model_name=MODEL_NAME)


def pad_embedding(values: np.ndarray, target_dimensions: int) -> List[float]:
    embedding = values.astype(np.float32).tolist()

    if target_dimensions <= 0:
        return embedding
    if len(embedding) > target_dimensions:
        raise RuntimeError(
            f"model output dimensions {len(embedding)} exceed target dimensions {target_dimensions}"
        )
    if len(embedding) < target_dimensions:
        embedding.extend([0.0] * (target_dimensions - len(embedding)))

    return embedding


def get_base_dimensions() -> int:
    model = get_model()
    sample = next(model.embed(["passage: dimension probe"]))
    return int(sample.shape[0])


def embed_text(text: str) -> EmbeddingResult:
    normalized = normalize_text(text)
    if not normalized:
        raise ValueError("text must not be empty")

    model = get_model()
    embedding = next(model.embed([to_passage_text(normalized)]))
    padded = pad_embedding(embedding, TARGET_DIMENSIONS)

    return EmbeddingResult(
        embedding=padded,
        dimensions=len(padded),
        base_dimensions=get_base_dimensions(),
        model=MODEL_NAME,
    )
