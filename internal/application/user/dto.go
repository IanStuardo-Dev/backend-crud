package userapp

type CreateInput struct {
	CompanyID       *int64
	Name            string
	Email           string
	Role            string
	IsActive        bool
	DefaultBranchID *int64
	Password        string
}

type UpdateInput struct {
	ID              int64
	CompanyID       *int64
	Name            string
	Email           string
	Role            string
	IsActive        bool
	DefaultBranchID *int64
}

type Output struct {
	ID              int64
	CompanyID       *int64
	Name            string
	Email           string
	Role            string
	IsActive        bool
	DefaultBranchID *int64
}
