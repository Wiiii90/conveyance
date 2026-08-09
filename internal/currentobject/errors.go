package currentobject

import "errors"

var (
	ErrInvalidReference          = errors.New("invalid reference")
	ErrInvalidEnvelope           = errors.New("invalid envelope")
	ErrUnsupportedEnvelopeFormat = errors.New("unsupported envelope format")
	ErrConflict                  = errors.New("conflict")
	ErrCurrentObjectNotFound     = errors.New("current object not found")
	ErrForbidden                 = errors.New("forbidden")
	ErrPayloadTooLarge           = errors.New("payload too large")
	ErrUnavailable               = errors.New("unavailable")
)
