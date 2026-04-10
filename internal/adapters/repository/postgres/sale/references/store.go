package salereferencespg

import (
	"context"
	"database/sql"

	postgresshared "github.com/IanStuardo-Dev/backend-crud/internal/adapters/repository/postgres/shared"
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

func (s *Store) UserExists(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := postgresshared.Queryer(ctx, s.DB).QueryRowContext(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE id = $1
		)`,
		userID,
	).Scan(&exists)
	return exists, err
}
