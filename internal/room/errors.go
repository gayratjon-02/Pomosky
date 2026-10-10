package room

import "errors"

var (
	ErrUnknownEvent      = errors.New("room: unknown event")
	ErrInvalidTransition = errors.New("room: invalid transition")
	ErrNotImplemented    = errors.New("room: not implemented")
)
