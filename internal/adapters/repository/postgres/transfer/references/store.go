package transferreferencespg

import (
	"context"
	"database/sql"
	"errors"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
)

type Store struct {
	DB *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) BranchExists(ctx context.Context, companyID, branchID int64) (bool, error) {
	var exists bool
	err := postgresshared.Queryer(ctx, s.DB).QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM branches
			WHERE company_id = $1 AND id = $2
		)`,
		companyID,
		branchID,
	).Scan(&exists)
	return exists, err
}

func (s *Store) RequesterCanAct(ctx context.Context, companyID, userID int64) (bool, error) {
	userRecord, err := s.loadUser(ctx, userID)
	if err != nil {
		return false, err
	}
	if !userRecord.IsActive {
		return false, nil
	}
	if userRecord.Role == domainuser.RoleSuperAdmin {
		return true, nil
	}
	return userRecord.CompanyID != nil && *userRecord.CompanyID == companyID, nil
}

func (s *Store) SupervisorEligible(ctx context.Context, companyID, userID int64) (bool, error) {
	userRecord, err := s.loadUser(ctx, userID)
	if err != nil {
		return false, err
	}
	if !userRecord.IsActive {
		return false, nil
	}
	if userRecord.Role == domainuser.RoleSuperAdmin {
		return true, nil
	}
	if userRecord.CompanyID == nil || *userRecord.CompanyID != companyID {
		return false, nil
	}
	return userRecord.Role == domainuser.RoleCompanyAdmin || userRecord.Role == domainuser.RoleInventoryManager, nil
}

type userRecord struct {
	CompanyID *int64
	Role      string
	IsActive  bool
}

func (s *Store) loadUser(ctx context.Context, userID int64) (userRecord, error) {
	var (
		record        userRecord
		companyIDNull sql.NullInt64
	)

	err := postgresshared.Queryer(ctx, s.DB).QueryRowContext(
		ctx,
		`SELECT company_id, role, is_active
		FROM users
		WHERE id = $1`,
		userID,
	).Scan(&companyIDNull, &record.Role, &record.IsActive)
	if errors.Is(err, sql.ErrNoRows) {
		return userRecord{}, nil
	}
	if err != nil {
		return userRecord{}, err
	}
	if companyIDNull.Valid {
		record.CompanyID = &companyIDNull.Int64
	}
	return record, nil
}
