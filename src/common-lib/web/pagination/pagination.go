package pagination

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/command"
)

const (
	nextLink = `<%v?%v>; rel="next"`
	// prevLink = `<%v?%v>; rel="prev"`

	nextLinkHeader = "Link"
)

// Pageable: An interface which needs to be implemnted by an entity which needs pagination
type Pageable interface {
	// returns the uuid for the last record from the slice of records which
	//is the pageable entity
	UniqueVal() *string

	// returns the default sorting field value for the last record from the
	// slice of records which is the pageable entity
	SortingVal() *string
}

// PaginationParams is the struct defining the params expected for pagination
type PaginationParams struct {
	Limit  int
	Cursor string
	sortBy *string
}

// SortBy : Getter for sortBy
func (pp *PaginationParams) SortBy() *string {
	return pp.sortBy
}

// NewParams will return an initialized pagination params struct
func NewParams(limit int, cursor string, sortBy *string) *PaginationParams {
	return &PaginationParams{
		Limit:  limit,
		Cursor: cursor,
		sortBy: sortBy,
	}
}

// GetFilter : Gets the filter for pagination. This filter can be converted to SQL using a SQL Converter
func GetFilter(paginate *PaginationParams, defaultSortByField string, uniqueIDField string, prefixFilter *filter.Filter, cnv command.Converter,
	fieldMapper func(string) string) (*filter.Filter, error) {
	//just return the prefix filter if paginate is nil
	if paginate == nil {
		return prefixFilter, nil
	}

	//make order by filter
	of, sortBy, err := makeOrderByFilter(paginate, defaultSortByField, uniqueIDField, cnv, fieldMapper)
	if err != nil {
		return prefixFilter, err
	}

	//append the order by filter to prefix filter and return it if limit is zero or less
	if paginate.Limit <= 0 {
		if prefixFilter == nil {
			return of, nil
		}

		pf := prefixFilter.OrderBy(of)
		return &pf, nil
	}

	//handle pagination
	var (
		pageFilter filter.Filter
	)

	//make cursor filter
	if len(paginate.Cursor) != 0 {

		cf, err := makeCursorFilter(paginate, sortBy, uniqueIDField, cnv, fieldMapper)
		if err != nil {
			return nil, err
		}

		//add order by filter
		pageFilter = cf.OrderBy(of)

		//we would need to append using AND if we got a cursor
		pageFilter.ShouldAnd = true

	} else {
		//add order by filter
		pageFilter = *of
	}

	//make limit filter
	lf, err := makeLimitFilter(paginate, cnv)
	if err != nil {
		return nil, err
	}

	//add limit filter
	pageFilter = pageFilter.Limit(lf)

	//append page filter to prefix filter if available
	if prefixFilter != nil {

		//if page filter includes a cusror filter, this has to be ANDed to prefix filter
		if pageFilter.ShouldAnd {
			prefixFilter = cnv.AND(prefixFilter, &pageFilter)

			//the final filter needs to be ANDed to where if we got a prefix filter
			prefixFilter.ShouldAnd = true

		} else {

			//no cursor filter, so no ANDing
			*prefixFilter = prefixFilter.Add(&pageFilter)
		}
	} else {

		prefixFilter = &pageFilter
	}

	return prefixFilter, nil
}

// Paginate : Truncates the one extra record (limit + 1) returned by GetFilter. Also sets the next page link in the response header.
func Paginate(w http.ResponseWriter, r http.Request, records Pageable, limit int, customSortBy *string) error {

	if reflect.ValueOf(records).Kind() != reflect.Ptr {
		return fmt.Errorf("paginate exepects records to be a pointer to the slice")
	}

	//Getting more companies than the limit implies we have more pages to follow
	rv := reflect.ValueOf(records).Elem()
	length := rv.Len()

	if (limit > 0) && (length > limit) {
		// truncate the (limit + 1)st record which was fetched to check if more pages left
		rv.SetLen(length - 1)

		//append link in response header
		err := setNextLink(w, r, records, limit, transformSortBy(customSortBy))
		if err != nil {
			return fmt.Errorf("failed to set next page link with error: %v", err)
		}
	}
	return nil
}

// setNextLink : sets the next link for pagination in the response header
func setNextLink(w http.ResponseWriter, r http.Request, records Pageable, limit int, customSortBy *string) error {

	//customSortBy can consist of both a field and order (asc, desc), separated by a space.
	//we only need the sort field for the cursor.
	var sortField *string
	if customSortBy != nil {
		sortField = &strings.Split(*customSortBy, " ")[0]
	}

	//get encoded cursor
	cursor, err := makeCursor(records, sortField)
	if err != nil {
		return fmt.Errorf("failed to get cursor with error: %v", err)
	}

	//get next link
	nextLink, err := makeNextLink(w, r, limit, cursor, customSortBy)
	if err != nil {
		return fmt.Errorf("failed to make next link: %v", err)
	}

	//set next link
	w.Header().Set(nextLinkHeader, nextLink)

	return nil
}

// infer the protocol from the request
func protocol(r http.Request) string {
	return r.URL.Scheme
}

// makeCursor : returns an encoded cursor, given the unique id, last record, sort by
func makeCursor(entity Pageable, customSortBy *string) (string, error) {

	if entity == nil {
		return "", fmt.Errorf("nil entity provided to makeCursor")
	}

	var (
		cursor  string
		sortVal string
	)

	if customSortBy != nil {

		//get the last record
		val := reflect.ValueOf(entity).Elem()
		rec := val.Index(val.Len() - 1).Interface()

		//get a map from the struct to lookup a sortBy field dynamically
		eb, err := json.Marshal(rec)
		if err != nil {
			return "", fmt.Errorf("failed to marshal entity with error: %v", err.Error())
		}

		entityMap := make(map[string]interface{})
		err = json.Unmarshal(eb, &entityMap)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal entity with error: %v", err.Error())
		}

		// find the sort key in the struct map
		sortVal = fmt.Sprintf("%v", entityMap[*customSortBy])

	} else {

		//use the default sorting key
		sv := entity.SortingVal()
		sortVal = *sv
	}

	//make the next cursor
	uv := entity.UniqueVal()

	c := Cursor{
		UniqueID:    *uv,
		OrderingKey: sortVal,
	}

	//encode the cursor
	cursor, err := c.Encode()
	if err != nil {
		return "", fmt.Errorf("failed to encode cursor with error: %v", err.Error())
	}

	return cursor, nil
}

// makes order by filter
func makeOrderByFilter(paginate *PaginationParams, defaultSortByField string, uniqueIDField string, cnv command.Converter, fieldMapper func(string) string) (*filter.Filter, string, error) {

	//default to default sort by
	sortBy := defaultSortByField

	//set the sortBy to the natural ordering, if not explicitly specified
	if (paginate != nil) && (paginate.SortBy() != nil) {
		if strings.TrimSpace(*paginate.sortBy) == "" {
			return nil, "", fmt.Errorf("failed to create ORDER BY filter for empty sortBy: %+v", paginate)
		}

		sortBy = *transformSortBy(paginate.sortBy)
	}

	//insert the uniqueIDField into sortBy
	sortBy = addUniqueFieldToSortBy(sortBy, uniqueIDField)
	of, err := cnv.GetOrderByFilter(sortBy, fieldMapper)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create ORDER BY filter for sortBy: %v with error: %v", sortBy, err)
	}

	return of, sortBy, err
}

// transformSortBy accepts a sortBy string and formats it into
// a commma-delimited list of sortFields followed by a space and sortOrder
func transformSortBy(sortBy *string) *string {
	if sortBy == nil || strings.TrimSpace(*sortBy) == "" {
		return nil
	}

	// split the sortBy into sort fields
	fields := strings.Split(*sortBy, ",")

	// remove possible whitespace around each field
	for i, s := range fields {
		fields[i] = strings.TrimSpace(s)
	}

	// the final field may contain a sort order (asc, desc)
	// we remove it from the list of fields
	fieldWithSortOption := strings.Fields(fields[len(fields)-1])
	hasSortOption := len(fieldWithSortOption) > 1
	if hasSortOption {
		fields[len(fields)-1] = fieldWithSortOption[0]
	}

	// join list of fields separated by commas
	transformedSortBy := strings.Join(fields, ",")

	// append the sort order if provided
	if hasSortOption {
		sortOrder := strings.ToLower(fieldWithSortOption[1])

		if sortOrder == "asc" || sortOrder == "desc" {
			transformedSortBy = transformedSortBy + " " + sortOrder
		}
	}

	return &transformedSortBy
}

func addUniqueFieldToSortBy(sortBy, uniqueIDField string) string {
	//splits sortBy into two strings, the first a comma-delimited list of fields
	//and the second, optionally, the sort order.
	sortByOptions := strings.Split(sortBy, " ")

	//appends the uniqueField to the comma-delimited list of fields if it's not already present
	sortByWithUniqueField := sortByOptions[0]
	if !sortByContainsUniqueField(sortByOptions[0], uniqueIDField) {
		sortByWithUniqueField = sortByOptions[0] + "," + uniqueIDField
	}

	//re-appends the sortOrder if provided
	if len(sortByOptions) > 1 {
		sortByWithUniqueField = sortByWithUniqueField + " " + sortByOptions[1]
	}

	return sortByWithUniqueField
}

func sortByContainsUniqueField(sortByFields string, uniqueField string) bool {
	fields := strings.Split(sortByFields, ",")
	for _, field := range fields {
		if field == uniqueField {
			return true
		}
	}

	return false
}

// makes cursor filter
func makeCursorFilter(paginate *PaginationParams, sortBy string, uniqueIDField string, cnv command.Converter, fieldMapper func(string) string) (*filter.Filter, error) {

	if (paginate == nil) || ((paginate != nil) && (len(paginate.Cursor) == 0)) {
		return nil, fmt.Errorf("empty cursor. cannot make a filter:%+v", paginate)
	}

	c := &Cursor{}
	err := c.Decode(paginate.Cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode the cursor in request with error: %v", err)
	}

	//sortBy can consist of both a field and order (asc, desc), separated by a space.
	//we use the fields option for the command property and the order option for operator.
	sortOptions := strings.Split(sortBy, " ")
	sortFields := sortOptions[0]
	sortOrder := getSortOrder(sortOptions)

	//we're using the user specified sortBy in defining cursor.
	//The pagination logic can go for a toss if the caller specifies a diff
	//sortBy critera in subsequent requests.
	//To curb this, the next link should also have the same sortBy as the first request.
	//Filter company records greater than (companyName, companyID)

	cmd := command.New(sortFields, getOrderOperator(sortOrder), "")
	companyCursorFilter, err := cnv.DoForCommandWithValue(cmd, fieldMapper)
	if err != nil {
		return nil, fmt.Errorf("command: %+v | failed to create cursor filter with error:%v", cmd, err)
	}

	if sortFields == uniqueIDField {
		return companyCursorFilter.CopyWithNewVals(c.OrderingKey), nil
	}

	return companyCursorFilter.CopyWithNewVals(c.OrderingKey, c.UniqueID), nil

}

// gets sort order for sortBy query: asc or desc
func getSortOrder(sortOptions []string) string {
	order := "asc"
	if len(sortOptions) > 1 && sortOptions[1] == "desc" {
		order = "desc"
	}

	return order
}

// gets command operator that corresponds to sort order
func getOrderOperator(sortOrder string) string {
	operator := string(command.Gt)
	if sortOrder == "desc" {
		operator = string(command.Lt)
	}

	return operator
}

// makes limit filter
func makeLimitFilter(paginate *PaginationParams, cnv command.Converter) (*filter.Filter, error) {

	if (paginate == nil) || ((paginate != nil) && (paginate.Limit <= 0)) {
		return nil, fmt.Errorf("invalid paginate or limit:%+v", paginate)
	}

	//increment the limit by 1 to check if more pages are left after this request
	newLimit := paginate.Limit + 1

	lf, err := cnv.GetLimitFilter(newLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to create LIMIT filter for limit: %v with error: %v", paginate.Limit, err)
	}
	return lf, nil
}

// make the next link for pagination
func makeNextLink(w http.ResponseWriter, r http.Request, limit int, cursor string, sortBy *string) (string, error) {
	queryWithCursor, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return "", err
	}
	queryWithCursor.Set("cursor", cursor)
	return fmt.Sprintf(nextLink, r.URL.Path, queryWithCursor.Encode()), nil
}
