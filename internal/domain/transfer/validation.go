package transfer

import (
	"strings"
	"time"
)

type ValidationError struct {
	Field   string
	Message string
}

type TransitionError struct {
	Kind    string
	Message string
}

const (
	TransitionInvalidState = "invalid_state"
	TransitionForbidden    = "forbidden"
)

func (e ValidationError) Error() string {
	if e.Field == "" {
		return e.Message
	}

	return e.Field + ": " + e.Message
}

func (e TransitionError) Error() string {
	return e.Message
}

func (t *Transfer) Normalize() {
	t.Note = strings.TrimSpace(t.Note)
	if t.Status == "" {
		t.Status = StatusPendingApproval
	}
}

func (t Transfer) ValidateForCreate() error {
	if t.CompanyID <= 0 {
		return ValidationError{Field: "company_id", Message: "company_id must be greater than 0"}
	}
	if t.OriginBranchID <= 0 {
		return ValidationError{Field: "origin_branch_id", Message: "origin_branch_id must be greater than 0"}
	}
	if t.DestinationBranchID <= 0 {
		return ValidationError{Field: "destination_branch_id", Message: "destination_branch_id must be greater than 0"}
	}
	if t.OriginBranchID == t.DestinationBranchID {
		return ValidationError{Field: "destination_branch_id", Message: "destination_branch_id must be different from origin_branch_id"}
	}
	if t.RequestedByUserID <= 0 {
		return ValidationError{Field: "requested_by_user_id", Message: "requested_by_user_id must be greater than 0"}
	}
	if t.SupervisorUserID <= 0 {
		return ValidationError{Field: "supervisor_user_id", Message: "supervisor_user_id must be greater than 0"}
	}
	if t.SupervisorUserID == t.RequestedByUserID {
		return ValidationError{Field: "supervisor_user_id", Message: "supervisor_user_id must be different from requested_by_user_id"}
	}
	if len(t.Items) == 0 {
		return ValidationError{Field: "items", Message: "items must contain at least one product"}
	}

	seenProducts := make(map[int64]struct{}, len(t.Items))
	for _, item := range t.Items {
		if item.ProductID <= 0 {
			return ValidationError{Field: "items", Message: "items must contain valid product_id values"}
		}
		if item.Quantity <= 0 {
			return ValidationError{Field: "items", Message: "items must contain quantities greater than 0"}
		}
		if _, exists := seenProducts[item.ProductID]; exists {
			return ValidationError{Field: "items", Message: "items must not repeat the same product"}
		}
		seenProducts[item.ProductID] = struct{}{}
	}

	return nil
}

func (t *Transfer) Approve(actorUserID int64, now time.Time) error {
	if actorUserID <= 0 {
		return ValidationError{Field: "approved_by_user_id", Message: "approved_by_user_id must be greater than 0"}
	}
	if t.Status != StatusPendingApproval {
		return TransitionError{Kind: TransitionInvalidState, Message: "transfer can only be approved when pending approval"}
	}
	if actorUserID != t.SupervisorUserID {
		return TransitionError{Kind: TransitionForbidden, Message: "only the assigned supervisor can approve this transfer"}
	}
	if actorUserID == t.RequestedByUserID {
		return TransitionError{Kind: TransitionForbidden, Message: "requester cannot approve their own transfer"}
	}

	t.Status = StatusApproved
	t.ApprovedByUserID = int64Ptr(actorUserID)
	t.ApprovedAt = timePtr(now.UTC())

	return nil
}

func (t *Transfer) Dispatch(actorUserID int64, now time.Time) error {
	if actorUserID <= 0 {
		return ValidationError{Field: "dispatched_by_user_id", Message: "dispatched_by_user_id must be greater than 0"}
	}
	if t.Status != StatusApproved {
		return TransitionError{Kind: TransitionInvalidState, Message: "transfer can only be dispatched when approved"}
	}

	t.Status = StatusInTransit
	t.DispatchedByUserID = int64Ptr(actorUserID)
	t.DispatchedAt = timePtr(now.UTC())

	return nil
}

func (t *Transfer) Receive(actorUserID int64, now time.Time) error {
	if actorUserID <= 0 {
		return ValidationError{Field: "received_by_user_id", Message: "received_by_user_id must be greater than 0"}
	}
	if t.Status != StatusInTransit {
		return TransitionError{Kind: TransitionInvalidState, Message: "transfer can only be marked as received when in transit"}
	}

	t.Status = StatusReceived
	t.ReceivedByUserID = int64Ptr(actorUserID)
	t.ReceivedAt = timePtr(now.UTC())

	return nil
}

func (t *Transfer) Cancel(actorUserID int64, now time.Time) error {
	if actorUserID <= 0 {
		return ValidationError{Field: "cancelled_by_user_id", Message: "cancelled_by_user_id must be greater than 0"}
	}
	switch t.Status {
	case StatusReceived:
		return TransitionError{Kind: TransitionInvalidState, Message: "received transfers cannot be cancelled"}
	case StatusCancelled:
		return TransitionError{Kind: TransitionInvalidState, Message: "transfer is already cancelled"}
	}

	t.Status = StatusCancelled
	t.CancelledByUserID = int64Ptr(actorUserID)
	t.CancelledAt = timePtr(now.UTC())

	return nil
}

func int64Ptr(value int64) *int64 {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
