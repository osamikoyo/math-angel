package errors

import "errors"

var (
	ErrEmptyTask    = errors.New("empty task")
	ErrAlreadyExist = errors.New("task already exist")
	ErrUnknown      = errors.New("unknown error")
	ErrNotFound     = errors.New("not found")
	ErrBadUID       = errors.New("bad uid")

	ErrEmptyQuery = errors.New("empty query")
)

// Validation errors
var (
	ErrEmptyUID     = errors.New("empty uid")
	ErrInvalidUID   = errors.New("unvalid uid")
	ErrEmptyProblem = errors.New("empty problem")
	ErrInvalidLevel = errors.New("invalid level")
)
