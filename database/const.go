package database

import "errors"

var (
	ErrArgInvalid      = errors.New("invalid argument")
	ErrExist           = errors.New("item already exists")
	ErrNoEvents        = errors.New("the times of pop event is more than push")
	ErrNotExist        = errors.New("item does not exist")
	ErrBigIntSetString = errors.New("big int setString error")
	ErrOutOfMemory     = errors.New("out of memory")
	ErrUnKnown         = errors.New("")
)
