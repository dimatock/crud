package crud

import (
	"fmt"
	"strings"
)

// Dialect defines the interface for database-specific SQL generation.
type Dialect interface {
	Placeholder(idx int) string
	InsertSQL(tableName string, cols, placeholders []string) string
	UpdateSQL(tableName string, setClauses string, pkColumn string, pkPlaceholder string) string
	SelectSQL(tableName string, cols []string, joins, whereClause, orderByClause, lockClause string, limit, offset int) string
	DeleteSQL(tableName string, pkColumn string, pkPlaceholder string) string
	UpsertSQL(tableName string, pkColumn string, cols []string) string
}

// DefaultSelectSQL provides a default implementation for building a SELECT query.
func DefaultSelectSQL(tableName string, cols []string, joins, whereClause, orderByClause, lockClause string, limit, offset int) string {
	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(cols, ", "), tableName)
	if joins != "" {
		sql += " " + joins
	}
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}
	if orderByClause != "" {
		sql += " ORDER BY " + orderByClause
	}
	if limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", offset)
	}
	if lockClause != "" {
		sql += " " + lockClause
	}
	return sql
}

// MySQLDialect implements Dialect for MySQL.
type MySQLDialect struct{}

func (d MySQLDialect) Placeholder(idx int) string {
	return "?"
}

func (d MySQLDialect) InsertSQL(tableName string, cols, placeholders []string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
}

func (d MySQLDialect) UpdateSQL(tableName string, setClauses string, pkColumn string, pkPlaceholder string) string {
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s", tableName, setClauses, pkColumn, pkPlaceholder)
}

func (d MySQLDialect) SelectSQL(tableName string, cols []string, joins, whereClause, orderByClause, lockClause string, limit, offset int) string {
	return DefaultSelectSQL(tableName, cols, joins, whereClause, orderByClause, lockClause, limit, offset)
}

func (d MySQLDialect) DeleteSQL(tableName string, pkColumn string, pkPlaceholder string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = %s", tableName, pkColumn, pkPlaceholder)
}

func (d MySQLDialect) UpsertSQL(tableName string, pkColumn string, cols []string) string {
	placeholders := make([]string, len(cols))
	updateClauses := make([]string, 0, len(cols))
	for i, col := range cols {
		placeholders[i] = "?"
		if col != pkColumn {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = VALUES(%s)", col, col))
		}
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		strings.Join(updateClauses, ", "),
	)
}

// SQLiteDialect implements Dialect for SQLite.
type SQLiteDialect struct{}

func (d SQLiteDialect) Placeholder(idx int) string {
	return "?"
}

func (d SQLiteDialect) InsertSQL(tableName string, cols, placeholders []string) string {
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
}

func (d SQLiteDialect) UpdateSQL(tableName string, setClauses string, pkColumn string, pkPlaceholder string) string {
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = %s", tableName, setClauses, pkColumn, pkPlaceholder)
}

func (d SQLiteDialect) SelectSQL(tableName string, cols []string, joins, whereClause, orderByClause, lockClause string, limit, offset int) string {
	return DefaultSelectSQL(tableName, cols, joins, whereClause, orderByClause, lockClause, limit, offset)
}

func (d SQLiteDialect) DeleteSQL(tableName string, pkColumn string, pkPlaceholder string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = %s", tableName, pkColumn, pkPlaceholder)
}

func (d SQLiteDialect) UpsertSQL(tableName string, pkColumn string, cols []string) string {
	placeholders := make([]string, len(cols))
	updateClauses := make([]string, 0, len(cols))
	for i, col := range cols {
		placeholders[i] = "?"
		if col != pkColumn {
			updateClauses = append(updateClauses, fmt.Sprintf("%s = excluded.%s", col, col))
		}
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT(%s) DO UPDATE SET %s",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
		pkColumn,
		strings.Join(updateClauses, ", "),
	)
}
