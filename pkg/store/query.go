package store

import (
	sq "github.com/Masterminds/squirrel"
)

// NewQuery creates a new query
func NewQuery() sq.SelectBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return sq.SelectBuilder(psql)
}

// NewDelete creates a new deletion query
func NewDelete() sq.DeleteBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return sq.DeleteBuilder(psql)
}

// Q is an alias for NewQuery
func Q() sq.SelectBuilder {
	return NewQuery()
}
