// Package filter represents the output of the query filter
// to a specific query language conversion. It allows to store
// a query and a value.
package filter

// Filter : stores the output of filter to query conversion.
// No fields are exported to avoid allowing multithreading errors where two threads mutate the same filter.
// Use Copy to copy a filter while changing it's values
type Filter struct {
	query  string
	values []interface{}

	// ShouldAnd : tells if the dynamic filter should be ANDed to a query or not.
	// This should be set to false when appending filters with just LIMIT, ORDER BY
	ShouldAnd bool
}

// New : creates a new Filter.
func New(q string, v ...interface{}) *Filter {
	filter := &Filter{
		query:     q,
		values:    nil,
		ShouldAnd: true,
	}

	if v != nil {
		filter.values = v

		return filter
	}

	return filter
}

// Add : Adds filter to f.
func (f Filter) Add(filter *Filter) Filter {
	if filter != nil {
		f.query = f.query + " " + filter.query
		f.values = append(f.values, filter.values...)
	}

	return f
}

// Limit : Appends LIMIT filter to f.
func (f Filter) Limit(limitFilter *Filter) Filter {
	if limitFilter != nil {
		f.query = f.query + " " + limitFilter.query

		for i := 0; i < len(limitFilter.values); i++ {
			f.values = append(f.values, limitFilter.values[i])
		}
	}

	return f
}

// OrderBy : Appends ORDER BY filter to f.
func (f Filter) OrderBy(orderBy *Filter) Filter {
	if orderBy != nil {
		f.query = f.query + " " + orderBy.query

		for i := 0; i < len(orderBy.values); i++ {
			f.values = append(f.values, orderBy.values[i])
		}
	}

	return f
}

// Get the SQL query.
// There is no SetQuery to avoid multithreading issues where one thread calls SetQuery on a filter used by another thread.
func (f Filter) GetQuery() string {
	return f.query
}

// Get the values that will be used during the query instead of the placeholders.
// There is no SetValues to avoid multithreading issues where one thread calls SetValues on a filter used by another thread.
// Instead, use .CopyWithNewVals to create a new filter that has the same query, but different values
func (f Filter) GetValues() []interface{} {
	return f.values
}

// Create a new filter with the same query, but with different values substituted
func (f Filter) CopyWithNewVals(values ...interface{}) *Filter {
	return &Filter{query: f.query, values: values, ShouldAnd: f.ShouldAnd}
}
