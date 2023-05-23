// Package filter houses the logic to handle
// complex filter in REST query parameter.
package filter

import (
	"context"
	"net/http"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/command"
)

const (
	queryFilter            = "filter"
	filterCtxKey filterCtx = "filter"
)

const (
	errGetFilter string = "operation: GetFilter | Typecasting error | filter: %w"
)

type filterCtx string

// Strategy : A strategy to parse a filter query
type Strategy interface {
	Parse(cnv command.Converter, query string, mapper func(string) string) (*filter.Filter, error)
}

// HandlerFunc : A type representing rest filter handler.
type HandlerFunc func(http.ResponseWriter, *http.Request, Strategy, command.Converter, func(string) string) (*http.Request, error)

// Middleware : Middleware wraps HTTP requests for complex filter handling.
// It parses the filter query parameter and sets a few headers in the request
// to be used in querying database.
func Middleware(w http.ResponseWriter, r *http.Request, st Strategy, cnv command.Converter, mapper func(string) string) (*http.Request, error) {
	vars := r.URL.Query()
	filter := vars.Get(queryFilter)

	if filter == "" {
		return r, nil
	}

	result, err := st.Parse(cnv, filter, mapper)
	if err != nil {
		//nolint:wrapcheck
		return r, err
	}

	//if the request already has a filter
	// prefix that filter to the user supplied filter
	f, err := GetFilter(r)
	if err != nil {
		//nolint:wrapcheck
		return r, err
	}

	if f != nil {
		result = cnv.AND(f, result)
	}

	r = SetRequestContext(result, r)

	return r, nil
}

// SetRequestContext : sets the filter into HTTP request context.
func SetRequestContext(f *filter.Filter, req *http.Request) *http.Request {

	ctx := context.WithValue(req.Context(), filterCtxKey, f)
	req = req.WithContext(ctx)
	return req
}

// GetFilter : helps get filter from request context.
func GetFilter(req *http.Request) (*filter.Filter, error) {

	if req.Context().Value(filterCtxKey) != nil {
		return req.Context().Value(filterCtxKey).(*filter.Filter), nil
	}
	return nil, nil
}
