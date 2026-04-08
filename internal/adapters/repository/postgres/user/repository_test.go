package postgresuser

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	userapp "github.com/IanStuardo-Dev/backend-crud/internal/application/user"
	domainuser "github.com/IanStuardo-Dev/backend-crud/internal/domain/user"
	"github.com/lib/pq"
)

func TestRepositoryCreateAssignsID(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	user := &domainuser.User{
		CompanyID:       int64Pointer(1),
		Name:            "Alice",
		Email:           "alice@example.com",
		Role:            domainuser.RoleCompanyAdmin,
		IsActive:        true,
		DefaultBranchID: int64Pointer(1),
		PasswordHash:    "hash",
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (company_id,name,email,role,is_active,default_branch_id,password_hash) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id")).
		WithArgs(int64(1), "Alice", "alice@example.com", domainuser.RoleCompanyAdmin, true, int64(1), "hash").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.ID != 5 {
		t.Fatalf("expected ID 5, got %d", user.ID)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryCreateReturnsConflictOnUniqueViolation(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	user := &domainuser.User{
		CompanyID:       int64Pointer(1),
		Name:            "Alice",
		Email:           "alice@example.com",
		Role:            domainuser.RoleCompanyAdmin,
		IsActive:        true,
		DefaultBranchID: int64Pointer(1),
		PasswordHash:    "hash",
	}

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO users (company_id,name,email,role,is_active,default_branch_id,password_hash) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id")).
		WithArgs(int64(1), "Alice", "alice@example.com", domainuser.RoleCompanyAdmin, true, int64(1), "hash").
		WillReturnError(&pq.Error{Code: "23505"})

	err := repo.Create(context.Background(), user)
	if !errors.Is(err, userapp.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryListReturnsUsers(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,company_id,name,email,role,is_active,default_branch_id,password_hash FROM users ORDER BY id")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "company_id", "name", "email", "role", "is_active", "default_branch_id", "password_hash"}).
			AddRow(1, 1, "Alice", "alice@example.com", domainuser.RoleCompanyAdmin, true, 1, "hash-1").
			AddRow(2, 1, "Bob", "bob@example.com", domainuser.RoleSalesUser, true, 1, "hash-2"))

	users, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Email != "alice@example.com" || users[1].Email != "bob@example.com" {
		t.Fatalf("unexpected users: %#v", users)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryGetByIDReturnsNilWhenMissing(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,company_id,name,email,role,is_active,default_branch_id,password_hash FROM users WHERE id=$1")).
		WithArgs(int64(10)).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got %#v", user)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryGetByIDReturnsUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,company_id,name,email,role,is_active,default_branch_id,password_hash FROM users WHERE id=$1")).
		WithArgs(int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "company_id", "name", "email", "role", "is_active", "default_branch_id", "password_hash"}).
			AddRow(3, 1, "Carol", "carol@example.com", domainuser.RoleInventoryManager, true, 1, "hash-3"))

	user, err := repo.GetByID(context.Background(), 3)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if user == nil || user.Email != "carol@example.com" {
		t.Fatalf("unexpected user: %#v", user)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryGetByEmailReturnsUser(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id,company_id,name,email,role,is_active,default_branch_id,password_hash FROM users WHERE email=$1")).
		WithArgs("carol@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "company_id", "name", "email", "role", "is_active", "default_branch_id", "password_hash"}).
			AddRow(3, 1, "Carol", "carol@example.com", domainuser.RoleInventoryManager, true, 1, "hash-3"))

	user, err := repo.GetByEmail(context.Background(), "carol@example.com")
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if user == nil || user.PasswordHash != "hash-3" {
		t.Fatalf("unexpected user: %#v", user)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryUpdateReturnsNotFoundWhenNoRowsAffected(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	user := &domainuser.User{ID: 7, CompanyID: int64Pointer(1), Name: "Alice", Email: "alice@example.com", Role: domainuser.RoleCompanyAdmin, IsActive: true, DefaultBranchID: int64Pointer(1)}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET company_id=$1, name=$2, email=$3, role=$4, is_active=$5, default_branch_id=$6 WHERE id=$7")).
		WithArgs(int64(1), "Alice", "alice@example.com", domainuser.RoleCompanyAdmin, true, int64(1), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Update(context.Background(), user)
	if !errors.Is(err, userapp.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryUpdateReturnsConflictOnUniqueViolation(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	user := &domainuser.User{ID: 7, CompanyID: int64Pointer(1), Name: "Alice", Email: "alice@example.com", Role: domainuser.RoleCompanyAdmin, IsActive: true, DefaultBranchID: int64Pointer(1)}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET company_id=$1, name=$2, email=$3, role=$4, is_active=$5, default_branch_id=$6 WHERE id=$7")).
		WithArgs(int64(1), "Alice", "alice@example.com", domainuser.RoleCompanyAdmin, true, int64(1), int64(7)).
		WillReturnError(&pq.Error{Code: "23505"})

	err := repo.Update(context.Background(), user)
	if !errors.Is(err, userapp.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryUpdateSucceeds(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)
	user := &domainuser.User{ID: 4, CompanyID: int64Pointer(1), Name: "Bob", Email: "bob@example.com", Role: domainuser.RoleSalesUser, IsActive: true, DefaultBranchID: int64Pointer(1)}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET company_id=$1, name=$2, email=$3, role=$4, is_active=$5, default_branch_id=$6 WHERE id=$7")).
		WithArgs(int64(1), "Bob", "bob@example.com", domainuser.RoleSalesUser, true, int64(1), int64(4)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Update(context.Background(), user); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryDeleteReturnsNotFoundWhenNoRowsAffected(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id=$1")).
		WithArgs(int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.Delete(context.Background(), 8)
	if !errors.Is(err, userapp.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	assertMockExpectations(t, mock)
}

func TestRepositoryDeleteSucceeds(t *testing.T) {
	db, mock := newMockDB(t)
	defer db.Close()

	repo := NewRepository(db)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id=$1")).
		WithArgs(int64(8)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.Delete(context.Background(), 8); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	assertMockExpectations(t, mock)
}

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	return db, mock
}

func assertMockExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet mock expectations: %v", err)
	}
}

func int64Pointer(value int64) *int64 {
	return &value
}
