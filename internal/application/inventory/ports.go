package inventoryapp

import (
	"context"

	domaininventory "github.com/IanStuardo-Dev/backend-crud/internal/domain/inventory"
)

type Repository interface {
	ListByBranch(ctx context.Context, companyID, branchID int64) ([]domaininventory.Item, error)
	SuggestSources(ctx context.Context, companyID, destinationBranchID, productID int64, quantity int) ([]domaininventory.SourceCandidate, error)
}

type UseCase interface {
	ListByBranch(ctx context.Context, input ListByBranchInput) ([]ItemOutput, error)
	SuggestSources(ctx context.Context, input SuggestSourcesInput) ([]SourceCandidateOutput, error)
}
