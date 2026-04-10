from concurrent import futures

import grpc

from embedding_runtime import GRPC_PORT, embed_text
from proto.embedding.v1 import embedding_pb2, embedding_pb2_grpc


class EmbeddingService(embedding_pb2_grpc.EmbeddingServiceServicer):
    def Embed(self, request, context):
        try:
            result = embed_text(request.text)
        except ValueError as exc:
            context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(exc))
        except Exception as exc:
            context.abort(grpc.StatusCode.INTERNAL, str(exc))

        return embedding_pb2.EmbedResponse(
            embedding=result.embedding,
            dimensions=result.dimensions,
            base_dimensions=result.base_dimensions,
            model=result.model,
        )


def start_grpc_server() -> grpc.Server:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=4))
    embedding_pb2_grpc.add_EmbeddingServiceServicer_to_server(EmbeddingService(), server)
    bound_port = server.add_insecure_port(f"[::]:{GRPC_PORT}")
    if bound_port == 0:
        raise RuntimeError(f"unable to bind gRPC server on port {GRPC_PORT}")

    server.start()
    return server
