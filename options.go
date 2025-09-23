package crud

import (
	"context"
	"fmt"
	"strings"
)

// Option configures a query.
type Option[T any] interface {
	apply(qb *queryBuilder[T]) error
}

// queryBuilder is an internal helper to construct SQL queries and hold relation-loading info.
type queryBuilder[T any] struct {
	dialect        Dialect // Reference to the dialect for placeholder generation
	whereClauses   []string
	joinClauses    []string
	orderByClauses []string
	lockClause     string // For row-locking clauses like FOR UPDATE
	limit          int
	offset         int
	args           []any
	relations      []Relation[T] // Holds relationship loading configurations
}

// --- Filter Option ---
type filterOption[T any] struct {
	column string
	value  any
}

func (o filterOption[T]) apply(qb *queryBuilder[T]) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s = %s", o.column, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// WithFilter adds a simple WHERE clause to the query (e.g., WHERE column = value).
func WithFilter[T any](column string, value any) Option[T] {
	return filterOption[T]{column: column, value: value}
}

// --- Operator Option ---
type operatorOption[T any] struct {
	column   string
	operator string
	value    any
}

func (o operatorOption[T]) apply(qb *queryBuilder[T]) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s %s %s", o.column, o.operator, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// WithOperator adds a WHERE clause with a custom operator (e.g., WHERE column > value).
func WithOperator[T any](column, operator string, value any) Option[T] {
	return operatorOption[T]{column: column, operator: operator, value: value}
}

// --- In Option ---
type inOption[T any] struct {
	column string
	values []any
}

func (o inOption[T]) apply(qb *queryBuilder[T]) error {
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
func WithIn[T any](column string, values ...any) Option[T] {
	return inOption[T]{column: column, values: values}
}

// --- Like Option ---
type likeOption[T any] struct {
	column string
	value  any
}

func (o likeOption[T]) apply(qb *queryBuilder[T]) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s LIKE %s", o.column, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// WithLike adds a WHERE LIKE clause to the query.
func WithLike[T any](column string, value any) Option[T] {
	return likeOption[T]{column: column, value: value}
}

// --- Lock Option ---
type lockOption[T any] struct {
	clause string
}

func (o lockOption[T]) apply(qb *queryBuilder[T]) error {
	qb.lockClause = o.clause
	return nil
}

// WithLock adds a row-locking clause to the query (e.g., "FOR UPDATE").
// This should only be used within a transaction.
func WithLock[T any](clause string) Option[T] {
	return lockOption[T]{clause: clause}
}

// --- Sort Option ---
type sortOption[T any] struct {
	column    string
	direction SortDirection
}

func (o sortOption[T]) apply(qb *queryBuilder[T]) error {
	qb.orderByClauses = append(qb.orderByClauses, fmt.Sprintf("%s %s", o.column, o.direction))
	return nil
}

// WithSort adds an ORDER BY clause to the query.
func WithSort[T any](column string, direction SortDirection) Option[T] {
	return sortOption[T]{column: column, direction: direction}
}

// --- Limit Option ---
type limitOption[T any] struct {
	limit int
}

func (o limitOption[T]) apply(qb *queryBuilder[T]) error {
	qb.limit = o.limit
	return nil
}

// WithLimit adds a LIMIT clause to the query.
func WithLimit[T any](limit int) Option[T] {
	return limitOption[T]{limit: limit}
}

// --- Offset Option ---
type offsetOption[T any] struct {
	offset int
}

func (o offsetOption[T]) apply(qb *queryBuilder[T]) error {
	qb.offset = o.offset
	return nil
}

// WithOffset adds an OFFSET clause to the query.
func WithOffset[T any](offset int) Option[T] {
	return offsetOption[T]{offset: offset}
}

// --- Join Option ---
type joinOption[T any] struct {
	joinClause string
}

func (o joinOption[T]) apply(qb *queryBuilder[T]) error {
	qb.joinClauses = append(qb.joinClauses, o.joinClause)
	return nil
}

// WithJoin adds a JOIN clause to the query (e.g., "INNER JOIN roles ON roles.id = users.role_id").
// IMPORTANT: When using joins, ensure column names in WithFilter and WithSort are fully qualified (e.g., "users.name").
func WithJoin[T any](joinClause string) Option[T] {
	return joinOption[T]{joinClause: joinClause}
}

// --- Subquery Option ---
type subqueryOption[T any] struct {
	column   string
	operator string
	subquery string
	args     []any
}

func (o subqueryOption[T]) apply(qb *queryBuilder[T]) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s %s (%s)", o.column, o.operator, o.subquery))
	qb.args = append(qb.args, o.args...)
	return nil
}

// WithSubquery adds a subquery clause (e.g., "id IN (SELECT user_id FROM ...)").
func WithSubquery[T any](column, operator, subquery string, args ...any) Option[T] {
	return subqueryOption[T]{column: column, operator: operator, subquery: subquery, args: args}
}

// --- Where Option ---
type whereOption[T any] struct {
	clause string
	args   []any
}

func (o whereOption[T]) apply(qb *queryBuilder[T]) error {
	qb.whereClauses = append(qb.whereClauses, o.clause)
	qb.args = append(qb.args, o.args...)
	return nil
}

// WithWhere adds a raw SQL WHERE clause. The user is responsible for writing the correct SQL and providing the correct arguments.
// The placeholders must match the dialect.
// Example: WithWhere("name = ? OR email = ?", "John", "john@example.com")
func WithWhere[T any](clause string, args ...any) Option[T] {
	return whereOption[T]{clause: clause, args: args}
}

// --- Eager Loading Options ---

// RelatedFetcher is a function type that fetches related entities for a given set of parent keys.
// K is the key type (e.g., int64, string).
// RelatedT is the type of the related entity.
type RelatedFetcher[K comparable, RelatedT any] func(ctx context.Context, keys []K) ([]RelatedT, error)

// relationOption is a generic Option for eager loading.
type relationOption[T any] struct {
	relation Relation[T]
}

// apply adds the relation to the queryBuilder.
func (o relationOption[T]) apply(qb *queryBuilder[T]) error {
	qb.relations = append(qb.relations, o.relation)
	return nil
}

// With adds a relationship to be eager-loaded.
// The provided mapper must implement the Relation interface.
func With[T any](mapper Relation[T]) Option[T] {
	return relationOption[T]{relation: mapper}
}
