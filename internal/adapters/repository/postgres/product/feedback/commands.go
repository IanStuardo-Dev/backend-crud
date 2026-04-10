package productfeedbackpg

import (
	"context"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	productapp "github.com/IanStuardo-Dev/backend-crud/internal/application/product"
	productdto "github.com/IanStuardo-Dev/backend-crud/internal/application/product/dto"
)

func (s *Store) SaveNeighborFeedback(ctx context.Context, input productdto.RecordNeighborFeedbackInput) (productdto.NeighborFeedbackOutput, error) {
	var output productdto.NeighborFeedbackOutput

	err := s.DB.QueryRowContext(
		ctx,
		`INSERT INTO product_neighbor_feedback (
			company_id,
			branch_id,
			source_product_id,
			suggested_product_id,
			user_id,
			action,
			note,
			created_at,
			updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())
		ON CONFLICT (company_id, branch_id, user_id, source_product_id, suggested_product_id)
		DO UPDATE SET
			action = EXCLUDED.action,
			note = EXCLUDED.note,
			updated_at = NOW()
		RETURNING source_product_id, suggested_product_id, company_id, branch_id, user_id, action, note, created_at, updated_at`,
		input.CompanyID,
		input.BranchID,
		input.SourceProductID,
		input.SuggestedProductID,
		input.UserID,
		input.Action,
		input.Note,
	).Scan(
		&output.SourceProductID,
		&output.SuggestedProductID,
		&output.CompanyID,
		&output.BranchID,
		&output.UserID,
		&output.Action,
		&output.Note,
		&output.CreatedAt,
		&output.UpdatedAt,
	)
	if postgresshared.IsForeignKeyViolation(err) {
		return productdto.NeighborFeedbackOutput{}, productapp.ErrInvalidReference
	}
	if err != nil {
		return productdto.NeighborFeedbackOutput{}, err
	}

	return output, nil
}
