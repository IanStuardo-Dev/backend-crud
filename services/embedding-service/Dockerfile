FROM python:3.11-slim

WORKDIR /app

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV PIP_NO_CACHE_DIR=1
ENV HF_HOME=/models/huggingface
ENV TRANSFORMERS_CACHE=/models/huggingface
ENV PORT=8000
ENV GRPC_PORT=50051
ENV EMBEDDING_MODEL=sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2
ENV EMBEDDING_TARGET_DIMENSIONS=1536
ENV EMBEDDING_DEVICE=cpu

COPY requirements.txt /app/requirements.txt

RUN pip install --no-cache-dir -r /app/requirements.txt

COPY app.py /app/app.py
COPY embedding_runtime.py /app/embedding_runtime.py
COPY grpc_service.py /app/grpc_service.py
COPY proto /app/proto
COPY text_normalization.py /app/text_normalization.py

EXPOSE 8000
EXPOSE 50051

CMD ["sh", "-c", "uvicorn app:app --host 0.0.0.0 --port ${PORT}"]
