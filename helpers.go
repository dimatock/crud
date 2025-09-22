package crud

import (
	"reflect"
)

// scannable represents an object that can be scanned, like *sql.Row or *sql.Rows.
type scannable interface {
	Scan(dest ...any) error
}

// scanRow scans a single row from the database into a new instance of T.
func (r *Repository[T]) scanRow(s scannable) (T, error) {
	var instance T
	destVal := reflect.ValueOf(&instance).Elem()
	scanDest := make([]any, len(r.columns))

	for i, col := range r.columns {
		if fieldIndex, ok := r.scanMap[col]; ok {
			scanDest[i] = destVal.Field(fieldIndex).Addr().Interface()
		} else {
			var dummyDest any
			scanDest[i] = &dummyDest
		}
	}

	if err := s.Scan(scanDest...); err != nil {
		return instance, err
	}
	return instance, nil
}
