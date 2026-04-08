package authhttp

import authapp "github.com/example/crud/internal/application/auth"

func toLoginInput(request loginRequest) authapp.LoginInput {
	return authapp.LoginInput{
		Email:    request.Email,
		Password: request.Password,
	}
}

func toLoginResponse(output authapp.LoginOutput) loginResponse {
	return loginResponse{
		AccessToken: output.AccessToken,
		TokenType:   output.TokenType,
		ExpiresAt:   output.ExpiresAt,
		User: userResponse{
			ID:              output.User.ID,
			CompanyID:       cloneInt64Pointer(output.User.CompanyID),
			Name:            output.User.Name,
			Email:           output.User.Email,
			Role:            output.User.Role,
			IsActive:        output.User.IsActive,
			DefaultBranchID: cloneInt64Pointer(output.User.DefaultBranchID),
		},
	}
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
