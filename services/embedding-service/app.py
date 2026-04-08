import os
from functools import lru_cache
from typing import Any, List

import numpy as np
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from text_normalization import normalize_text, to_passage_text


MODEL_NAME = os.getenv("EMBEDDING_MODEL", "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2")
TARGET_DIMENSIONS = int(os.getenv("EMBEDDING_TARGET_DIMENSIONS", "1536"))
DEVICE = os.getenv("EMBEDDING_DEVICE", "cpu")

app = FastAPI(title="local-semantic-embedding-service", version="1.0.0")


class EmbedRequest(BaseModel):
    text: str = Field(min_length=1)


class EmbedResponse(BaseModel):
    embedding: List[float]
    dimensions: int
    base_dimensions: int
    model: str


class HealthResponse(BaseModel):
    status: str
    model: str
    device: str
    target_dimensions: int
    base_dimensions: int


@lru_cache
def get_model() -> Any:
    from fastembed import TextEmbedding

    return TextEmbedding(model_name=MODEL_NAME)

def pad_embedding(values: np.ndarray, target_dimensions: int) -> List[float]:
    embedding = values.astype(np.float32).tolist()

    if target_dimensions <= 0:
        return embedding
    if len(embedding) > target_dimensions:
        raise HTTPException(
            status_code=500,
            detail=f"model output dimensions {len(embedding)} exceed target dimensions {target_dimensions}",
        )
    if len(embedding) < target_dimensions:
        embedding.extend([0.0] * (target_dimensions - len(embedding)))

    return embedding


def get_base_dimensions() -> int:
    model = get_model()
    sample = next(model.embed(["passage: dimension probe"]))
    return int(sample.shape[0])


@app.get("/health", response_model=HealthResponse)
def health() -> HealthResponse:
    get_model()
    base_dimensions = get_base_dimensions()

    return HealthResponse(
        status="ok",
        model=MODEL_NAME,
        device=DEVICE,
        target_dimensions=TARGET_DIMENSIONS,
        base_dimensions=base_dimensions,
    )


@app.post("/embed", response_model=EmbedResponse)
def embed(request: EmbedRequest) -> EmbedResponse:
    normalized = normalize_text(request.text)
    if not normalized:
        raise HTTPException(status_code=422, detail="text must not be empty")

    model = get_model()
    embedding = next(model.embed([to_passage_text(normalized)]))
    padded = pad_embedding(embedding, TARGET_DIMENSIONS)

    return EmbedResponse(
        embedding=padded,
        dimensions=len(padded),
        base_dimensions=get_base_dimensions(),
        model=MODEL_NAME,
    )
