package crud

import (
	"fmt"
	"strings"
)

// Option configures a query.
type Option interface {
	apply(qb *queryBuilder) error
}

// queryBuilder is an internal helper to construct SQL queries.
type queryBuilder struct {
	dialect        Dialect // Reference to the dialect for placeholder generation
	whereClauses   []string
	joinClauses    []string
	orderByClauses []string
	lockClause     string // For row-locking clauses like FOR UPDATE
	limit          int
	offset         int
	args           []any
}

// --- Filter Option ---
type filterOption struct {
	column string
	value  any
}

func (o filterOption) apply(qb *queryBuilder) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s = %s", o.column, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// WithFilter adds a simple WHERE clause to the query (e.g., WHERE column = value).
func WithFilter(column string, value any) Option {
	return filterOption{column: column, value: value}
}

// --- Operator Option ---
type operatorOption struct {
	column   string
	operator string
	value    any
}

func (o operatorOption) apply(qb *queryBuilder) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s %s %s", o.column, o.operator, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// WithOperator adds a WHERE clause with a custom operator (e.g., WHERE column > value).
func WithOperator(column, operator string, value any) Option {
	return operatorOption{column: column, operator: operator, value: value}
}

// --- In Option ---
type inOption struct {
	column string
	values []any
}

func (o inOption) apply(qb *queryBuilder) error {
	if len(o.values) == 0 {
		return fmt.Errorf("WithIn option requires at least one value for column '%s'", o.column)
	}
	placeholders := make([]string, len(o.values))
	for i := range o.values {
		placeholders[i] = qb.dialect.Placeholder(len(qb.args) + 1 + i)
	}
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s IN (%s)", o.column, strings.Join(placeholders, ",")))
	qb.args = append(qb.args, o.values...)
	return nil
}

// WithIn adds a WHERE IN clause to the query.
func WithIn(column string, values ...any) Option {
	return inOption{column: column, values: values}
}

// --- Like Option ---
type likeOption struct {
	column string
	value  any
}

func (o likeOption) apply(qb *queryBuilder) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s LIKE %s", o.column, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// WithLike adds a WHERE LIKE clause to the query.
func WithLike(column string, value any) Option {
	return likeOption{column: column, value: value}
}

// --- Lock Option ---
type lockOption struct {
	clause string
}

func (o lockOption) apply(qb *queryBuilder) error {
	qb.lockClause = o.clause
	return nil
}

// WithLock adds a row-locking clause to the query (e.g., "FOR UPDATE").
// This should only be used within a transaction.
func WithLock(clause string) Option {
	return lockOption{clause: clause}
}

// --- Sort Option ---
type sortOption struct {
	column    string
	direction SortDirection
}

func (o sortOption) apply(qb *queryBuilder) error {
	qb.orderByClauses = append(qb.orderByClauses, fmt.Sprintf("%s %s", o.column, o.direction))
	return nil
}

// WithSort adds an ORDER BY clause to the query.
func WithSort(column string, direction SortDirection) Option {
	return sortOption{column: column, direction: direction}
}

// --- Limit Option ---
type limitOption struct {
	limit int
}

func (o limitOption) apply(qb *queryBuilder) error {
	qb.limit = o.limit
	return nil
}

// WithLimit adds a LIMIT clause to the query.
func WithLimit(limit int) Option {
	return limitOption{limit: limit}
}

// --- Offset Option ---
type offsetOption struct {
	offset int
}

func (o offsetOption) apply(qb *queryBuilder) error {
	qb.offset = o.offset
	return nil
}

// WithOffset adds an OFFSET clause to the query.
func WithOffset(offset int) Option {
	return offsetOption{offset: offset}
}

// --- Join Option ---
type joinOption struct {
	joinClause string
}

func (o joinOption) apply(qb *queryBuilder) error {
	qb.joinClauses = append(qb.joinClauses, o.joinClause)
	return nil
}

// WithJoin adds a JOIN clause to the query (e.g., "INNER JOIN roles ON roles.id = users.role_id").
// IMPORTANT: When using joins, ensure column names in WithFilter and WithSort are fully qualified (e.g., "users.name").
func WithJoin(joinClause string) Option {
	return joinOption{joinClause: joinClause}
}

// --- Subquery Option ---
type subqueryOption struct {
	column   string
	operator string
	subquery string
	args     []any
}

func (o subqueryOption) apply(qb *queryBuilder) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s %s (%s)", o.column, o.operator, o.subquery))
	qb.args = append(qb.args, o.args...)
	return nil
}

// WithSubquery adds a subquery clause (e.g., "id IN (SELECT user_id FROM ...)").
func WithSubquery(column, operator, subquery string, args ...any) Option {
	return subqueryOption{column: column, operator: operator, subquery: subquery, args: args}
}

// --- Where Option ---
type whereOption struct {
	clause string
	args   []any
}

func (o whereOption) apply(qb *queryBuilder) error {
	qb.whereClauses = append(qb.whereClauses, o.clause)
	qb.args = append(qb.args, o.args...)
	return nil
}

// WithWhere adds a raw SQL WHERE clause. The user is responsible for writing the correct SQL and providing the correct arguments.
// The placeholders must match the dialect.
// Example: WithWhere("name = ? OR email = ?", "John", "john@example.com")
func WithWhere(clause string, args ...any) Option {
	return whereOption{clause: clause, args: args}
}