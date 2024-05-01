package cursorpager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// CursorPaginationAttribute represent the cursor pagination response.
type CursorPaginationAttribute struct {
	NextCursor string `json:"next_cursor"`
	PrevCursor string `json:"prev_cursor"`
}

// preCursor represents the cursor.
// After converting this structure into byte slices, the base64-decoded value becomes the cursor string.
type preCursor struct {
	// valid represents whether the cursor is valid or not.
	// Used only for identification in the package.
	valid bool `json:"-"`
	// CursorID represents the ID of the cursor.
	// This value must always be unique within the listing.
	// Since the type also changes, it is defined as any type, and since the cursor argument in Query-type methods passes this value, type assertions must be made,
	// but since functions can be passed, the behavior in case of failure can also be customized.
	CursorID any `json:"id"`
	// CursorPointsNext represents the direction of the cursor.
	// If true, the cursor points to the next data.
	// If false, the cursor points to the previous data.
	CursorPointsNext bool `json:"points_next"`
	// SubCursorName represents the name of the sub-cursor.
	// It is necessary to indicate which sort order was used.
	SubCursorName string `json:"sub_cursor_name"`
	// SubCursor represents the value of the sub-cursor.
	// The value will vary depending on what was adopted in the sorting order.
	// Since the type also changes, it is defined as any type, and since the SubCursorValue argument in Query-type methods passes this value,
	// it is necessary to make a type assertion, but since a function can be passed, the behavior in case of failure can also be customized.
	SubCursor any `json:"sub_cursor"`
}

func createPreCursor(id any, pointsNext bool, name string, value any) preCursor {
	c := preCursor{
		valid:            true,
		CursorID:         id,
		CursorPointsNext: pointsNext,
		SubCursorName:    name,
		SubCursor:        value,
	}
	return c
}

func generatePager(next, prev preCursor) CursorPaginationAttribute {
	return CursorPaginationAttribute{
		NextCursor: encodeCursor(next),
		PrevCursor: encodeCursor(prev),
	}
}

func encodeCursor(cursor preCursor) string {
	if !cursor.valid {
		return ""
	}
	serializedCursor, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	encodedCursor := base64.StdEncoding.EncodeToString(serializedCursor)
	return encodedCursor
}

func decodeCursor(cursor string) (preCursor, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return preCursor{}, ErrFailedDecodeCursor
	}

	var cur preCursor
	if err := json.Unmarshal(decodedCursor, &cur); err != nil {
		return preCursor{}, ErrFailedDecodeCursor
	}
	return cur, nil
}

type cursorData struct {
	ID    any
	Name  string
	Value any
}

func calculatePagination(
	isFirstPage, hasPagination, pointsNext bool, firstData, lastData cursorData,
) CursorPaginationAttribute {
	pagination := CursorPaginationAttribute{}
	nextCur := preCursor{}
	prevCur := preCursor{}
	if isFirstPage {
		if hasPagination {
			nextCur := createPreCursor(lastData.ID, true, lastData.Name, lastData.Value)
			pagination = generatePager(nextCur, preCursor{})
		}
	} else {
		if pointsNext {
			// if pointing next, it always has prev but it might not have next
			if hasPagination {
				nextCur = createPreCursor(lastData.ID, true, lastData.Name, lastData.Value)
			}
			prevCur = createPreCursor(firstData.ID, false, firstData.Name, firstData.Value)
			pagination = generatePager(nextCur, prevCur)
		} else {
			// this is case of prev, there will always be nest, but prev needs to be calculated
			nextCur = createPreCursor(lastData.ID, true, lastData.Name, lastData.Value)
			if hasPagination {
				prevCur = createPreCursor(firstData.ID, false, firstData.Name, firstData.Value)
			}
			pagination = generatePager(nextCur, prevCur)
		}
	}
	return pagination
}

// GetCursorData retrieves data with cursor pagination.
func GetCursorData[T any](
	q Querier[T],
	cursor string,
	order OrderMethod,
	limit int32,
) ([]T, CursorPaginationAttribute, error) {
	var err error
	isFirst := cursor == "" // is this the first request?
	pointsNext := false     // is the cursor pointing to the next data?
	SubCursor := order.GetCursorKeyName()
	var decodedCursor preCursor
	var cursorValue any
	var data []T
	if !isFirst {
		cursorCheck := func(cur string) bool {
			decodedCursor, err = decodeCursor(cur)
			if err != nil {
				return false
			}
			if decodedCursor.SubCursorName != SubCursor {
				return false
			}
			cursorValue = decodedCursor.SubCursor
			return true
		}
		if !cursorCheck(cursor) {
			isFirst = true
		}
	}

	if !isFirst {
		// Take over the direction of the specified cursor this time
		pointsNext = decodedCursor.CursorPointsNext
		var cursorDir string
		if pointsNext {
			cursorDir = "next"
		} else {
			cursorDir = "prev"
		}
		ID := decodedCursor.CursorID
		data, err = q.RunQueryWithCursorParamsFunc(SubCursor, order.GetStringValue(), limit+1, cursorDir, ID, cursorValue)
		if err != nil {
			return nil, CursorPaginationAttribute{}, fmt.Errorf("failed to run query with cursor params: %w", err)
		}
	} else {
		data, err = q.RunQueryWithLimitFunc(order.GetStringValue(), limit+1)
		if err != nil {
			return nil, CursorPaginationAttribute{}, fmt.Errorf("failed to run query with numbered params: %w", err)
		}
	}

	if len(data) == 0 { // case of data has no record
		return nil, CursorPaginationAttribute{}, ErrDataNoRecord
	}
	hasPagination := len(data) > int(limit)
	if hasPagination {
		data = data[:limit]
	}
	eLen := len(data)

	var firstValue, lastValue any
	lastIndex := eLen - 1
	if lastIndex < 0 {
		lastIndex = 0
	}
	firstID, firstValue := q.CursorIDAndValueSelector(SubCursor, data[0])
	lastID, lastValue := q.CursorIDAndValueSelector(SubCursor, data[lastIndex])

	firstData := cursorData{
		ID:    firstID,
		Name:  SubCursor,
		Value: firstValue,
	}
	lastData := cursorData{
		ID:    lastID,
		Name:  SubCursor,
		Value: lastValue,
	}
	var pageInfo CursorPaginationAttribute
	if pointsNext || isFirst {
		// No cursor specified or if the direction is next, calculate in the same order
		pageInfo = calculatePagination(isFirst, hasPagination, pointsNext, firstData, lastData)
	} else {
		// If the direction is prev, calculate in the reverse order
		// For example, if the data is 1, 2, 3, 4, 5, 6,....
		// if limit is 2, cursor is 5, and pointsNext is false(prev),
		// (To make it easier to understand, after getting the data of 1 and 2, we get the data of 3 and 4 with next,
		// and when the data of 5 and 6 are obtained by next again, prev becomes 5 and next becomes 6. This is a case where there is a request for prev at this time.)
		// Even if the order specification is in ascending order, the reverse order is used for prev, so firstData becomes 4 and lastData becomes 3.
		// If the pagination is calculated as it is, 4 will be in prev and 3 in next,
		// The next access to next will return 4 and 5, which are the next two values after 3, even though 5 and 6, which are the next two values after 4, are expected.
		// Also, when you access prev, it returns 3, 2, which are two before 4, even though it expects 2, 1, which are two before 3.
		pageInfo = calculatePagination(isFirst, hasPagination, pointsNext, lastData, firstData)
	}

	return data, pageInfo, nil
}
