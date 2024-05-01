package cursorpager

// OrderMethod indicates sort order
type OrderMethod interface {
	// GetCursorKeyName returns the associated cursor key
	GetCursorKeyName() string
	// GetStringValue returns the string representation of the order
	GetStringValue() string
}
