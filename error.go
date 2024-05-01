package cursorpager

import "errors"

var (
	// ErrDataNoRecord represents the error that the target data does not exist.
	ErrDataNoRecord = errors.New("cursor pagination target data does not exist")

	// ErrFailedDecodeCursor represents the error that the cursor decoding failed.
	ErrFailedDecodeCursor = errors.New("failed to decode cursor")
)
