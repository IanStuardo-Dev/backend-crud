package user

const (
	RoleSuperAdmin       = "super_admin"
	RoleCompanyAdmin     = "company_admin"
	RoleInventoryManager = "inventory_manager"
	RoleSalesUser        = "sales_user"
)

// User represents the core user entity.
type User struct {
	ID              int64
	CompanyID       *int64
	Name            string
	Email           string
	Role            string
	IsActive        bool
	DefaultBranchID *int64
	PasswordHash    string
}
