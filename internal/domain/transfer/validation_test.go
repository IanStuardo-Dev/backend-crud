package transfer

import (
	"errors"
	"testing"
	"time"
)

func TestTransferValidateForCreateRequiresSupervisorDifferentFromRequester(t *testing.T) {
	transfer := Transfer{
		CompanyID:           1,
		OriginBranchID:      1,
		DestinationBranchID: 2,
		RequestedByUserID:   7,
		SupervisorUserID:    7,
		Items: []Item{{
			ProductID: 10,
			Quantity:  1,
		}},
	}

	var validationErr ValidationError
	err := transfer.ValidateForCreate()
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if validationErr.Field != "supervisor_user_id" {
		t.Fatalf("expected supervisor_user_id validation, got %#v", validationErr)
	}
}

func TestTransferApproveRequiresAssignedSupervisor(t *testing.T) {
	now := time.Now().UTC()
	transfer := Transfer{
		Status:            StatusPendingApproval,
		RequestedByUserID: 7,
		SupervisorUserID:  9,
	}

	var transitionErr TransitionError
	err := transfer.Approve(10, now)
	if !errors.As(err, &transitionErr) {
		t.Fatalf("expected TransitionError, got %v", err)
	}
	if transitionErr.Kind != TransitionForbidden {
		t.Fatalf("expected forbidden transition, got %#v", transitionErr)
	}
}

func TestTransferDispatchAndReceiveAdvanceStatus(t *testing.T) {
	now := time.Now().UTC()
	transfer := Transfer{
		Status:            StatusApproved,
		RequestedByUserID: 7,
		SupervisorUserID:  9,
	}

	if err := transfer.Dispatch(8, now); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if transfer.Status != StatusInTransit || transfer.DispatchedByUserID == nil || *transfer.DispatchedByUserID != 8 {
		t.Fatalf("unexpected transfer after dispatch: %#v", transfer)
	}

	if err := transfer.Receive(11, now.Add(time.Hour)); err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if transfer.Status != StatusReceived || transfer.ReceivedByUserID == nil || *transfer.ReceivedByUserID != 11 {
		t.Fatalf("unexpected transfer after receive: %#v", transfer)
	}
}
