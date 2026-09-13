package service

import (
	apperr "upcycle-hub/pkg/errors"
)

func ErrValidation(msg string) error {
	return apperr.New(apperr.CodeValidation, msg)
}

func ErrForbidden(msg string) error {
	return apperr.New(apperr.CodeForbidden, msg)
}

func ErrNotFound(msg string) error {
	return apperr.New(apperr.CodeNotFound, msg)
}
