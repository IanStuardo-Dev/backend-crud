package postgresuser

import (
	"context"
	"database/sql"
	"errors"

	userapp "github.com/IanStuardo-Dev/backend-crud/internal/application/user"
	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
	"github.com/lib/pq"
)

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) userapp.Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *domainuser.User) error {
	err := r.db.QueryRowContext(
		ctx,
		"INSERT INTO users (company_id,name,email,role,is_active,default_branch_id,password_hash) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id",
		nullableInt64(user.CompanyID),
		user.Name,
		user.Email,
		user.Role,
		user.IsActive,
		nullableInt64(user.DefaultBranchID),
		user.PasswordHash,
	).Scan(&user.ID)
	if isUniqueViolation(err) {
		return userapp.ErrConflict
	}

	return err
}

func (r *repository) List(ctx context.Context) ([]domainuser.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id,company_id,name,email,role,is_active,default_branch_id,password_hash FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domainuser.User, 0)
	for rows.Next() {
		var user domainuser.User
		if err := scanUser(rows, &user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*domainuser.User, error) {
	var user domainuser.User
	err := scanUser(
		r.db.QueryRowContext(ctx, "SELECT id,company_id,name,email,role,is_active,default_branch_id,password_hash FROM users WHERE id=$1", id),
		&user,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var user domainuser.User
	err := scanUser(
		r.db.QueryRowContext(ctx, "SELECT id,company_id,name,email,role,is_active,default_branch_id,password_hash FROM users WHERE email=$1", email),
		&user,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *repository) Update(ctx context.Context, user *domainuser.User) error {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE users SET company_id=$1, name=$2, email=$3, role=$4, is_active=$5, default_branch_id=$6 WHERE id=$7",
		nullableInt64(user.CompanyID),
		user.Name,
		user.Email,
		user.Role,
		user.IsActive,
		nullableInt64(user.DefaultBranchID),
		user.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return userapp.ErrConflict
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return userapp.ErrNotFound
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return userapp.ErrNotFound
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner, user *domainuser.User) error {
	var (
		companyID       sql.NullInt64
		defaultBranchID sql.NullInt64
	)
	if err := row.Scan(
		&user.ID,
		&companyID,
		&user.Name,
		&user.Email,
		&user.Role,
		&user.IsActive,
		&defaultBranchID,
		&user.PasswordHash,
	); err != nil {
		return err
	}
	user.CompanyID = toOptionalInt64(companyID)
	user.DefaultBranchID = toOptionalInt64(defaultBranchID)
	return nil
}

func toOptionalInt64(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	cloned := value.Int64
	return &cloned
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
