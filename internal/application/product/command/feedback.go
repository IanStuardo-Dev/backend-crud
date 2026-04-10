package productcommand

import (
	"context"
	"strings"

	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
	producterrors "github.com/IanStuardo-Dev/backend-crud/internal/application/product/errors"
	productports "github.com/IanStuardo-Dev/backend-crud/internal/application/product/ports"
	domainproduct "github.com/IanStuardo-Dev/backend-crud/internal/domain/product"
)

type RecordFeedbackHandler struct {
	reader productports.ProductCatalogReader
	writer productports.ProductSuggestionFeedbackWriter
}

func NewRecordFeedbackHandler(reader productports.ProductCatalogReader, writer productports.ProductSuggestionFeedbackWriter) RecordFeedbackHandler {
	return RecordFeedbackHandler{reader: reader, writer: writer}
}

func (h RecordFeedbackHandler) Handle(ctx context.Context, input productdto.RecordNeighborFeedbackInput) (productdto.NeighborFeedbackOutput, error) {
	if err := validateID(input.SourceProductID); err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}
	if err := validateNeighborFeedbackID("suggested_product_id", input.SuggestedProductID); err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}
	if err := validateNeighborFeedbackID("branch_id", input.BranchID); err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}
	if err := validateNeighborFeedbackID("user_id", input.UserID); err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}
	if input.SourceProductID == input.SuggestedProductID {
		return productdto.NeighborFeedbackOutput{}, domainproduct.ValidationError{
			Field:   "suggested_product_id",
			Message: "suggested_product_id must be different from the source product",
		}
	}

	action, err := normalizeNeighborFeedbackAction(input.Action)
	if err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}
	note, err := normalizeNeighborFeedbackNote(input.Note)
	if err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}

	sourceProduct, err := h.reader.GetByID(ctx, input.SourceProductID)
	if err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}
	if sourceProduct == nil {
		return productdto.NeighborFeedbackOutput{}, producterrors.ErrNotFound
	}

	suggestedProduct, err := h.reader.GetByID(ctx, input.SuggestedProductID)
	if err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}
	if suggestedProduct == nil {
		return productdto.NeighborFeedbackOutput{}, producterrors.ErrNotFound
	}
	if sourceProduct.CompanyID != suggestedProduct.CompanyID {
		return productdto.NeighborFeedbackOutput{}, producterrors.ErrInvalidReference
	}

	return h.writer.SaveNeighborFeedback(ctx, productdto.RecordNeighborFeedbackInput{
		SourceProductID:    input.SourceProductID,
		SuggestedProductID: input.SuggestedProductID,
		CompanyID:          sourceProduct.CompanyID,
		BranchID:           input.BranchID,
		UserID:             input.UserID,
		Action:             action,
		Note:               note,
	})
}

func validateNeighborFeedbackID(field string, id int64) error {
	if id <= 0 {
		return domainproduct.ValidationError{Field: field, Message: field + " must be greater than 0"}
	}

	return nil
}

func normalizeNeighborFeedbackAction(value string) (string, error) {
	action := strings.ToLower(strings.TrimSpace(value))
	switch action {
	case "accepted", "rejected", "ignored":
		return action, nil
	default:
		return "", domainproduct.ValidationError{
			Field:   "action",
			Message: "action must be one of accepted, rejected, or ignored",
		}
	}
}

func normalizeNeighborFeedbackNote(value string) (string, error) {
	note := strings.TrimSpace(value)
	if len(note) > 1000 {
		return "", domainproduct.ValidationError{
			Field:   "note",
			Message: "note must be less than or equal to 1000 characters",
		}
	}

	return note, nil
}
