package auth

import "fmt"

type ErrorCode string

const (
	ErrInvalidPayload       ErrorCode = "invalid_payload"
	ErrMissingUUID          ErrorCode = "missing_uuid"
	ErrMissingUsername      ErrorCode = "missing_username"
	ErrMissingServerID      ErrorCode = "missing_server_id"
	ErrMissingProof         ErrorCode = "missing_proof"
	ErrInvalidChallenge     ErrorCode = "invalid_challenge"
	ErrWrongClient          ErrorCode = "challenge_wrong_client"
	ErrInvalidProof         ErrorCode = "invalid_proof"
	ErrAlreadyAuthenticated ErrorCode = "already_authenticated"
)

type AuthError struct {
	Code    ErrorCode
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}
