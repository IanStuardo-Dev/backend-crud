package productapp

import (
	"context"

	productcommand "github.com/IanStuardo-Dev/backend-crud/internal/application/product/command"
	productquery "github.com/IanStuardo-Dev/backend-crud/internal/application/product/query"
)

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (Output, error)
	List(ctx context.Context) ([]Output, error)
	GetByID(ctx context.Context, id int64) (Output, error)
	FindNeighbors(ctx context.Context, input FindNeighborsInput) (FindNeighborsOutput, error)
	RecordNeighborFeedback(ctx context.Context, input RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error)
	Update(ctx context.Context, input UpdateInput) (Output, error)
	Delete(ctx context.Context, id int64) error
}

type useCase struct {
	create         productcommand.CreateHandler
	list           productquery.ListHandler
	getByID        productquery.GetByIDHandler
	findNeighbors  productquery.FindNeighborsHandler
	recordFeedback productcommand.RecordFeedbackHandler
	update         productcommand.UpdateHandler
	delete         productcommand.DeleteHandler
}

func newUseCase(reader ProductCatalogReader, writer ProductCatalogWriter, similarity ProductSimilarityReader, feedback ProductSuggestionFeedbackWriter, embedder Embedder) UseCase {
	return &useCase{
		create:         productcommand.NewCreateHandler(writer, embedder),
		list:           productquery.NewListHandler(reader),
		getByID:        productquery.NewGetByIDHandler(reader),
		findNeighbors:  productquery.NewFindNeighborsHandler(reader, similarity),
		recordFeedback: productcommand.NewRecordFeedbackHandler(reader, feedback),
		update:         productcommand.NewUpdateHandler(reader, writer, embedder),
		delete:         productcommand.NewDeleteHandler(writer),
	}
}

func NewUseCase(args ...any) UseCase {
	switch len(args) {
	case 2:
		reader := args[0].(ProductCatalogReader)
		writer := args[0].(ProductCatalogWriter)
		similarity := args[0].(ProductSimilarityReader)
		feedback := args[0].(ProductSuggestionFeedbackWriter)
		embedder := extractEmbedder(args[1])
		return newUseCase(reader, writer, similarity, feedback, embedder)
	case 5:
		return newUseCase(
			args[0].(ProductCatalogReader),
			args[1].(ProductCatalogWriter),
			args[2].(ProductSimilarityReader),
			args[3].(ProductSuggestionFeedbackWriter),
			extractEmbedder(args[4]),
		)
	default:
		panic("invalid product use case dependencies")
	}
}

func extractEmbedder(value any) Embedder {
	if value == nil {
		return nil
	}
	return value.(Embedder)
}

func (uc *useCase) Create(ctx context.Context, input CreateInput) (Output, error) {
	return uc.create.Handle(ctx, input)
}

func (uc *useCase) List(ctx context.Context) ([]Output, error) {
	return uc.list.Handle(ctx)
}

func (uc *useCase) GetByID(ctx context.Context, id int64) (Output, error) {
	return uc.getByID.Handle(ctx, id)
}

func (uc *useCase) FindNeighbors(ctx context.Context, input FindNeighborsInput) (FindNeighborsOutput, error) {
	return uc.findNeighbors.Handle(ctx, input)
}

func (uc *useCase) RecordNeighborFeedback(ctx context.Context, input RecordNeighborFeedbackInput) (NeighborFeedbackOutput, error) {
	return uc.recordFeedback.Handle(ctx, input)
}

func (uc *useCase) Update(ctx context.Context, input UpdateInput) (Output, error) {
	return uc.update.Handle(ctx, input)
}

func (uc *useCase) Delete(ctx context.Context, id int64) error {
	return uc.delete.Handle(ctx, id)
}
