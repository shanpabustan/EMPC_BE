package errRbac

import "errors"

var (
	ErrResourceNotFound  = errors.New("resource not found")
	ErrResourceInUse     = errors.New("resource is in use")
	ErrResourceNameTaken = errors.New("resource name is already in use")
)
