package semanticgrpc

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
	grpcpb "github.com/IanStuardo-Dev/backend-crud/internal/infrastructure/embedding/grpcpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const defaultTimeout = 15 * time.Second

type Service struct {
	target      string
	timeout     time.Duration
	dialOptions []grpc.DialOption

	mu     sync.Mutex
	conn   *grpc.ClientConn
	client grpcpb.EmbeddingServiceClient
}

func NewService(target string, timeout time.Duration) *Service {
	return newService(target, timeout)
}

func newService(target string, timeout time.Duration, dialOptions ...grpc.DialOption) *Service {
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	if len(dialOptions) == 0 {
		dialOptions = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
	}

	return &Service{
		target:      strings.TrimSpace(target),
		timeout:     timeout,
		dialOptions: dialOptions,
	}
}

func (s *Service) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if s == nil || s.target == "" {
		return nil, fmt.Errorf("semantic embedding service is not configured")
	}

	rpcContext := ctx
	var cancel context.CancelFunc
	if s.timeout > 0 {
		rpcContext, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}

	client, err := s.getClient(rpcContext)
	if err != nil {
		return nil, err
	}

	response, err := client.Embed(rpcContext, &grpcpb.EmbedRequest{Text: text})
	if err != nil {
		return nil, formatRPCError(err)
	}
	if len(response.GetEmbedding()) != domainproduct.EmbeddingDimensions {
		return nil, fmt.Errorf("embedding must have exactly %d dimensions, got %d", domainproduct.EmbeddingDimensions, len(response.GetEmbedding()))
	}
	for index, value := range response.GetEmbedding() {
		if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
			return nil, fmt.Errorf("embedding contains an invalid value at position %d", index)
		}
	}

	return response.GetEmbedding(), nil
}

func (s *Service) Close() error {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return nil
	}

	err := s.conn.Close()
	s.conn = nil
	s.client = nil
	return err
}

func (s *Service) getClient(ctx context.Context) (grpcpb.EmbeddingServiceClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.client != nil {
		return s.client, nil
	}

	conn, err := grpc.DialContext(ctx, s.target, s.dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("connect embedding service: %w", err)
	}

	s.conn = conn
	s.client = grpcpb.NewEmbeddingServiceClient(conn)
	return s.client, nil
}

func formatRPCError(err error) error {
	rpcStatus, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("embedding service rpc failed: %w", err)
	}

	code := strings.ToLower(rpcStatus.Code().String())
	if rpcStatus.Code() == codes.OK {
		return nil
	}
	if message := strings.TrimSpace(rpcStatus.Message()); message != "" {
		return fmt.Errorf("embedding service rpc %s: %s", code, message)
	}

	return fmt.Errorf("embedding service rpc %s", code)
}
