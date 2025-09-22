package crud

import (
	"fmt"
	"strconv"
	"strings"
)

// PostgresDialect implements Dialect for PostgreSQL.
type PostgresDialect struct{}

// Placeholder returns the placeholder for the given index (e.g., $1, $2).
func (d PostgresDialect) Placeholder(idx int) string {
	return "$" + strconv.Itoa(idx)
}

// InsertSQL generates the INSERT statement for PostgreSQL.
func (d PostgresDialect) InsertSQL(tableName string, cols, placeholders []string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)
}

// UpdateSQL generates the UPDATE statement for PostgreSQL.
func (d PostgresDialect) UpdateSQL(tableName string, setClauses string, pkColumn string, pkPlaceholder string) string {
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s", tableName, setClauses, pkColumn, pkPlaceholder)
}

// SelectSQL generates the SELECT statement for PostgreSQL.
func (d PostgresDialect) SelectSQL(
	tableName string, cols []string, joins, whereClause string, orderByClause, lockClause string, limit, offset int,
) string {
	return DefaultSelectSQL(tableName, cols, joins, whereClause, orderByClause, lockClause, limit, offset)
}

// DeleteSQL generates the DELETE statement for PostgreSQL.
func (d PostgresDialect) DeleteSQL(tableName string, pkColumn string, pkPlaceholder string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = %s", tableName, pkColumn, pkPlaceholder)
}

// UpsertSQL generates the INSERT ... ON CONFLICT statement for PostgreSQL.
func (d PostgresDialect) UpsertSQL(tableName string, pkColumn string, cols []string) string {
	placeholders := make([]string, len(cols))
	updateClauses := make([]string, 0, len(cols))
	for i, col := range cols {
		placeholders[i] = d.Placeholder(i + 1)
		if col != pkColumn {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		pkColumn,
		strings.Join(updateClauses, ", "),
	)
}
