package errors

import "errors"


// General errors
var (
	ErrNotFound     = errors.New("not found")
	ErrUnknown      = errors.New("unknown error")
	ErrBadUID       = errors.New("bad uid")
	ErrAlreadyExist = errors.New("task already exist")
	ErrEmptyUID     = errors.New("empty uid")
	ErrInvalidUID   = errors.New("unvalid uid")
	ErrEmptyProblem = errors.New("empty problem")
	ErrInvalidLevel = errors.New("invalid level")
)

// Cache errors
var (
	ErrFailedMarshal      = errors.New("json marshal fail")
	ErrInternalCacheError = errors.New("internal cash error")
	ErrFailedDecode       = errors.New("decode cash data fail")
)

// Task errors
var (
	ErrEmptyTask    = errors.New("empty task")
	ErrEmptyQuery = errors.New("empty query")
)

// Solution errors
var (
	ErrEmptySolution = errors.New("empty solution")
)

// User service errors
var (
	ErrFailedHash              = errors.New("failed hashed")
	ErrEmptyUser               = errors.New("empty user")
	ErrEmptyUsernameOrPassword = errors.New("empty username or password")
	ErrAuthFailed              = errors.New("authentication error")
	ErrInvalidToken            = errors.New("invalid token")
)
