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

// Where adds a WHERE clause to the query. It is a flexible method that can handle
// different numbers of arguments to create different types of conditions:
//   - Where(column, value) for simple equality (e.g., "username", "john") -> WHERE username = ?
//   - Where(column, operator, value) for complex comparisons (e.g., "age", ">", 21) -> WHERE age > ?
//   - Where(rawClause, args...) for raw SQL (e.g., "status = ? OR archived = ?", "active", false)
func Where[T any](args ...any) Option[T] {
	if len(args) == 0 {
		return noOpOption[T]{}
	}

	clause, isClause := args[0].(string)
	if !isClause {
		// The first argument must be a string (column name or raw clause)
		return noOpOption[T]{}
	}

	// Case 1: Raw query. Check for '?' as a heuristic.
	if strings.Contains(clause, "?") {
		return rawWhereOption[T]{clause: clause, args: args[1:]}
	}

	// Case 2: Operator query (col, op, val)
	if len(args) == 3 {
		if operator, ok := args[1].(string); ok {
			return operatorWhereOption[T]{column: clause, operator: operator, value: args[2]}
		}
	}

	// Case 3: Simple equality (col, val)
	if len(args) == 2 {
		return simpleWhereOption[T]{column: clause, value: args[1]}
	}

	// If none of the above, it's an invalid combination
	return noOpOption[T]{}
}

// noOpOption is an option that does nothing. Used as a fallback for invalid Where args.
type noOpOption[T any] struct{}

func (o noOpOption[T]) apply(_ *queryBuilder[T]) error {
	return nil
}

// --- Simple Where Option (column = value) ---
type simpleWhereOption[T any] struct {
	column string
	value  any
}

func (o simpleWhereOption[T]) apply(qb *queryBuilder[T]) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s = %s", o.column, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// --- Operator Where Option (column op value) ---
type operatorWhereOption[T any] struct {
	column   string
	operator string
	value    any
}

func (o operatorWhereOption[T]) apply(qb *queryBuilder[T]) error {
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s %s %s", o.column, o.operator, qb.dialect.Placeholder(len(qb.args)+1)))
	qb.args = append(qb.args, o.value)
	return nil
}

// --- In Option ---
type inOption[T any] struct {
	column string
	values []any
}

func (o inOption[T]) apply(qb *queryBuilder[T]) error {
	if len(o.values) == 0 {
		return fmt.Errorf("WhereIn option requires at least one value for column '%s'", o.column)
	}
	placeholders := make([]string, len(o.values))
	for i := range o.values {
		placeholders[i] = qb.dialect.Placeholder(len(qb.args) + 1 + i)
	}
	qb.whereClauses = append(qb.whereClauses, fmt.Sprintf("%s IN (%s)", o.column, strings.Join(placeholders, ",")))
	qb.args = append(qb.args, o.values...)
	return nil
}

// WhereIn adds a WHERE IN clause to the query.
func WhereIn[T any](column string, values ...any) Option[T] {
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

// WhereLike adds a WHERE LIKE clause to the query.
func WhereLike[T any](column string, value any) Option[T] {
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

// Lock adds a row-locking clause to the query (e.g., "FOR UPDATE").
// This should only be used within a transaction.
func Lock[T any](clause string) Option[T] {
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

// OrderBy adds an ORDER BY clause to the query.
func OrderBy[T any](column string, direction SortDirection) Option[T] {
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

// Limit adds a LIMIT clause to the query.
func Limit[T any](limit int) Option[T] {
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

// Offset adds an OFFSET clause to the query.
func Offset[T any](offset int) Option[T] {
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

// Join adds a JOIN clause to the query (e.g., "INNER JOIN roles ON roles.id = users.role_id").
// IMPORTANT: When using joins, ensure column names in Where and OrderBy are fully qualified (e.g., "users.name").
func Join[T any](joinClause string) Option[T] {
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

// WhereSubquery adds a subquery clause (e.g., "id IN (SELECT user_id FROM ...)").
func WhereSubquery[T any](column, operator, subquery string, args ...any) Option[T] {
	return subqueryOption[T]{column: column, operator: operator, subquery: subquery, args: args}
}

// --- Raw Where Option (raw sql) ---
type rawWhereOption[T any] struct {
	clause string
	args   []any
}

func (o rawWhereOption[T]) apply(qb *queryBuilder[T]) error {
	// The number of arguments *before* this clause is added
	argStartIndex := len(qb.args)

	finalClause := ""
	argCounterForThisClause := 0
	for _, char := range o.clause {
		if char == '?' {
			// Use the global argument index
			globalArgIndex := argStartIndex + argCounterForThisClause
			finalClause += qb.dialect.Placeholder(globalArgIndex + 1) // Placeholder is 1-based
			argCounterForThisClause++
		} else {
			finalClause += string(char)
		}
	}

	if argCounterForThisClause != len(o.args) {
		return fmt.Errorf("mismatched number of placeholders (?) and arguments in Where clause: '%s'", o.clause)
	}

	qb.whereClauses = append(qb.whereClauses, finalClause)
	qb.args = append(qb.args, o.args...)
	return nil
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

// WithRelation adds a relationship to be eager-loaded.
// The provided mapper must implement the Relation interface.
func WithRelation[T any](mapper Relation[T]) Option[T] {
	return relationOption[T]{relation: mapper}
}
