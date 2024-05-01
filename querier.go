package cursorpager

// Querier is an interface that defines the query method.
type Querier[T any] interface {
	// RunQueryWithCursorParamsFunc executes a query with cursor parameters.
	RunQueryWithCursorParamsFunc(
		subCursor, orderMethod string, limit int32,
		cursorDir string, cursor, subCursorValue any,
	) ([]T, error)
	// RunQueryWithLimitFunc executes a query with limit parameters.
	// It is used only when the first page is displayed.
	RunQueryWithLimitFunc(orderMethod string, limit int32) ([]T, error)
	// CursorIDAndValueSelector selects the cursor ID and value.
	CursorIDAndValueSelector(subCursor string, e T) (any, any)
}
