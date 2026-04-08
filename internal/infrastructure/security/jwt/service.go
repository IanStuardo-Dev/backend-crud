package jwt

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authapp "github.com/IanStuardo-Dev/backend-crud/internal/application/auth"
)

type Service struct {
	secret []byte
	issuer string
	ttl    time.Duration
	now    func() time.Time
}

type claims struct {
	CompanyID       *int64 `json:"company_id,omitempty"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	IsActive        bool   `json:"is_active"`
	DefaultBranchID *int64 `json:"default_branch_id,omitempty"`
	jwt.RegisteredClaims
}

func NewService(secret, issuer string, ttl time.Duration) *Service {
	return &Service{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *Service) Generate(user authapp.AuthenticatedUser) (authapp.IssuedToken, error) {
	now := s.now().UTC()
	expiresAt := now.Add(s.ttl)
	tokenClaims := claims{
		CompanyID:       cloneInt64Pointer(user.CompanyID),
		Name:            user.Name,
		Email:           user.Email,
		Role:            user.Role,
		IsActive:        user.IsActive,
		DefaultBranchID: cloneInt64Pointer(user.DefaultBranchID),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return authapp.IssuedToken{}, err
	}

	return authapp.IssuedToken{
		AccessToken: signedToken,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *Service) Verify(token string) (authapp.AuthenticatedUser, error) {
	parsedClaims := &claims{}
	parsedToken, err := jwt.ParseWithClaims(token, parsedClaims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, authapp.ErrUnauthorized
		}
		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	if err != nil || !parsedToken.Valid {
		return authapp.AuthenticatedUser{}, authapp.ErrUnauthorized
	}

	userID, err := strconv.ParseInt(parsedClaims.Subject, 10, 64)
	if err != nil {
		return authapp.AuthenticatedUser{}, authapp.ErrUnauthorized
	}
	if userID <= 0 {
		return authapp.AuthenticatedUser{}, authapp.ErrUnauthorized
	}
	if parsedClaims.ExpiresAt == nil || parsedClaims.ExpiresAt.Time.Before(s.now()) {
		return authapp.AuthenticatedUser{}, authapp.ErrUnauthorized
	}

	return authapp.AuthenticatedUser{
		ID:              userID,
		CompanyID:       cloneInt64Pointer(parsedClaims.CompanyID),
		Name:            parsedClaims.Name,
		Email:           parsedClaims.Email,
		Role:            parsedClaims.Role,
		IsActive:        parsedClaims.IsActive,
		DefaultBranchID: cloneInt64Pointer(parsedClaims.DefaultBranchID),
	}, nil
}

func (s *Service) SetNow(now func() time.Time) {
	if now == nil {
		return
	}
	s.now = now
}

func IsUnauthorized(err error) bool {
	return errors.Is(err, authapp.ErrUnauthorized)
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
