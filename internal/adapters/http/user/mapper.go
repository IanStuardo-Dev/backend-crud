package userhttp

import userapp "github.com/IanStuardo-Dev/backend-crud/internal/application/user"

func toCreateInput(request createUserRequest) userapp.CreateInput {
	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	return userapp.CreateInput{
		CompanyID:       cloneInt64Pointer(request.CompanyID),
		Name:            request.Name,
		Email:           request.Email,
		Role:            request.Role,
		IsActive:        isActive,
		DefaultBranchID: cloneInt64Pointer(request.DefaultBranchID),
		Password:        request.Password,
	}
}

func toUpdateInput(id int64, request updateUserRequest) userapp.UpdateInput {
	return userapp.UpdateInput{
		ID:              id,
		CompanyID:       cloneInt64Pointer(request.CompanyID),
		Name:            request.Name,
		Email:           request.Email,
		Role:            request.Role,
		IsActive:        request.IsActive,
		DefaultBranchID: cloneInt64Pointer(request.DefaultBranchID),
	}
}

func toUserResponse(output userapp.Output) userResponse {
	return userResponse{
		ID:              output.ID,
		CompanyID:       cloneInt64Pointer(output.CompanyID),
		Name:            output.Name,
		Email:           output.Email,
		Role:            output.Role,
		IsActive:        output.IsActive,
		DefaultBranchID: cloneInt64Pointer(output.DefaultBranchID),
	}
}

func toUserResponses(outputs []userapp.Output) []userResponse {
	responses := make([]userResponse, 0, len(outputs))
	for _, output := range outputs {
		responses = append(responses, toUserResponse(output))
	}

	return responses
}

func cloneInt64Pointer(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
