package domain

import "errors"

var (
	ErrNotFound            = errors.New("resource not found")
	ErrUnauthorized        = errors.New("unauthorized access")
	ErrForbidden           = errors.New("forbidden: insufficient privileges")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrEmailAlreadyExists  = errors.New("email is already registered")
	ErrTenantInactive      = errors.New("tenant account is inactive or pending approval")
	ErrQuotaExceeded       = errors.New("quota limit exceeded: storage or transcription minutes exhausted")
	ErrIntegrityMismatch   = errors.New("cryptographic integrity hash mismatch")
	ErrInvalidSessionState = errors.New("invalid session state for this action")
	ErrInvalidFileFormat   = errors.New("unsupported or malicious media file format")
)

