package semanticgrpc

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
	grpcpb "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/embedding/grpcpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const testBufSize = 1024 * 1024

type testEmbeddingServer struct {
	grpcpb.UnimplementedEmbeddingServiceServer
	embedFn func(context.Context, *grpcpb.EmbedRequest) (*grpcpb.EmbedResponse, error)
}

func (s testEmbeddingServer) Embed(ctx context.Context, request *grpcpb.EmbedRequest) (*grpcpb.EmbedResponse, error) {
	return s.embedFn(ctx, request)
}

func TestServiceEmbedText(t *testing.T) {
	t.Run("returns embedding from service", func(t *testing.T) {
		service := newBufconnService(t, testEmbeddingServer{
			embedFn: func(_ context.Context, request *grpcpb.EmbedRequest) (*grpcpb.EmbedResponse, error) {
				if request.GetText() != "Cafe Marley" {
					t.Fatalf("unexpected text %q", request.GetText())
				}

				embedding := make([]float32, domainproduct.EmbeddingDimensions)
				for index := range embedding {
					embedding[index] = 0.001
				}

				return &grpcpb.EmbedResponse{
					Embedding:      embedding,
					Dimensions:     int32(len(embedding)),
					BaseDimensions: 384,
					Model:          "test-model",
				}, nil
			},
		})
		defer service.Close()

		embedding, err := service.EmbedText(context.Background(), "Cafe Marley")
		if err != nil {
			t.Fatalf("EmbedText() error = %v", err)
		}
		if len(embedding) != domainproduct.EmbeddingDimensions {
			t.Fatalf("expected %d dimensions, got %d", domainproduct.EmbeddingDimensions, len(embedding))
		}
	})

	t.Run("surfaces remote error detail", func(t *testing.T) {
		service := newBufconnService(t, testEmbeddingServer{
			embedFn: func(_ context.Context, _ *grpcpb.EmbedRequest) (*grpcpb.EmbedResponse, error) {
				return nil, status.Error(codes.Unavailable, "model is warming up")
			},
		})
		defer service.Close()

		_, err := service.EmbedText(context.Background(), "Cafe Marley")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "model is warming up") {
			t.Fatalf("expected remote detail, got %v", err)
		}
	})

	t.Run("rejects unexpected dimensions", func(t *testing.T) {
		service := newBufconnService(t, testEmbeddingServer{
			embedFn: func(_ context.Context, _ *grpcpb.EmbedRequest) (*grpcpb.EmbedResponse, error) {
				return &grpcpb.EmbedResponse{Embedding: []float32{0.1, 0.2}}, nil
			},
		})
		defer service.Close()

		_, err := service.EmbedText(context.Background(), "Cafe Marley")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "exactly") {
			t.Fatalf("expected dimensions error, got %v", err)
		}
	})
}

func newBufconnService(t *testing.T, server testEmbeddingServer) *Service {
	t.Helper()

	listener := bufconn.Listen(testBufSize)
	grpcServer := grpc.NewServer()
	grpcpb.RegisterEmbeddingServiceServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Logf("bufconn grpc server stopped: %v", err)
		}
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
		_ = listener.Close()
	})

	return newService(
		"bufnet",
		time.Second,
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
}
