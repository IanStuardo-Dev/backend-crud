package transfer

import "strings"

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}

	return e.Field + ": " + e.Message
}

func (t *Transfer) Normalize() {
	t.Note = strings.TrimSpace(t.Note)
}

func (t Transfer) ValidateForCreate() error {
	if t.CompanyID <= 0 {
		return ValidationError{Field: "company_id", Message: "company_id must be greater than 0"}
	}
	if t.OriginBranchID <= 0 {
		return ValidationError{Field: "origin_branch_id", Message: "origin_branch_id must be greater than 0"}
	}
	if t.DestinationBranchID <= 0 {
		return ValidationError{Field: "destination_branch_id", Message: "destination_branch_id must be greater than 0"}
	}
	if t.OriginBranchID == t.DestinationBranchID {
		return ValidationError{Field: "destination_branch_id", Message: "destination_branch_id must be different from origin_branch_id"}
	}
	if t.RequestedByUserID <= 0 {
		return ValidationError{Field: "requested_by_user_id", Message: "requested_by_user_id must be greater than 0"}
	}
	if len(t.Items) == 0 {
		return ValidationError{Field: "items", Message: "items must contain at least one product"}
	}

	seenProducts := make(map[int64]struct{}, len(t.Items))
	for _, item := range t.Items {
		if item.ProductID <= 0 {
			return ValidationError{Field: "items", Message: "items must contain valid product_id values"}
		}
		if item.Quantity <= 0 {
			return ValidationError{Field: "items", Message: "items must contain quantities greater than 0"}
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return ValidationError{Field: "items", Message: "items must not repeat the same product"}
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	return nil
}
