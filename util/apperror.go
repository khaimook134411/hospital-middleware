package util

import "net/http"

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string { return e.Message }

func newErr(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

func ErrValidation(msg string) *AppError {
	return newErr(http.StatusBadRequest, "VALIDATION_ERROR", msg)
}
func ErrUnauthorized() *AppError {
	return newErr(http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
}
func ErrInvalidCredentials() *AppError {
	return newErr(http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid credentials")
}
func ErrForbidden() *AppError {
	return newErr(http.StatusForbidden, "FORBIDDEN", "forbidden")
}
func ErrHospitalNotFound() *AppError {
	return newErr(http.StatusNotFound, "HOSPITAL_NOT_FOUND", "hospital not found")
}
func ErrUsernameTaken() *AppError {
	return newErr(http.StatusConflict, "USERNAME_TAKEN", "username already taken")
}
func ErrHISUnavailable() *AppError {
	return newErr(http.StatusBadGateway, "HIS_UNAVAILABLE", "HIS service unavailable")
}
func ErrInternal() *AppError {
	return newErr(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
