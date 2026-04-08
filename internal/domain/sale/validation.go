package sale

import "fmt"

// ValidationError represents a business validation failure.
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

func (s Sale) ValidateForCreate() error {
	if s.CompanyID <= 0 {
		return ValidationError{Field: "company_id", Message: "company_id must be greater than 0"}
	}
	if s.BranchID <= 0 {
		return ValidationError{Field: "branch_id", Message: "branch_id must be greater than 0"}
	}
	if s.CreatedByUserID <= 0 {
		return ValidationError{Field: "created_by_user_id", Message: "created_by_user_id must be greater than 0"}
	}
	if len(s.Items) == 0 {
		return ValidationError{Field: "items", Message: "items must contain at least one product"}
	}

	seenProducts := make(map[int64]struct{}, len(s.Items))
	for index, item := range s.Items {
		if item.ProductID <= 0 {
			return ValidationError{Field: fmt.Sprintf("items[%d].product_id", index), Message: "product_id must be greater than 0"}
		}
		if item.Quantity <= 0 {
			return ValidationError{Field: fmt.Sprintf("items[%d].quantity", index), Message: "quantity must be greater than 0"}
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return ValidationError{Field: "items", Message: "items must not repeat the same product"}
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	return nil
}
