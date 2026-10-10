package room

import "errors"

var (
	ErrUnknownEvent      = errors.New("room: unknown event")
	ErrInvalidEvent      = errors.New("room: invalid event data")
	ErrInvalidTransition = errors.New("room: invalid transition")
)
