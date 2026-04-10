from typing import List

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from embedding_runtime import DEVICE, MODEL_NAME, TARGET_DIMENSIONS, embed_text, get_base_dimensions, get_model
from grpc_service import start_grpc_server

app = FastAPI(title="local-semantic-embedding-service", version="1.0.0")
grpc_server = None


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


@app.on_event("startup")
def startup() -> None:
    global grpc_server

    get_model()
    grpc_server = start_grpc_server()


@app.on_event("shutdown")
def shutdown() -> None:
    if grpc_server is None:
        return

    grpc_server.stop(5).wait(5)


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
    try:
        result = embed_text(request.text)
    except ValueError as exc:
        raise HTTPException(status_code=422, detail=str(exc)) from exc
    except RuntimeError as exc:
        raise HTTPException(status_code=500, detail=str(exc)) from exc

    return EmbedResponse(
        embedding=result.embedding,
        dimensions=result.dimensions,
        base_dimensions=result.base_dimensions,
        model=result.model,
    )
