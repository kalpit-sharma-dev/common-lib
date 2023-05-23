package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/converters/sql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/pagination"
)

const (
	companyID   = "01d3b488-b123-4411-8cd4-7955d4b5a412"
	companyName = "avengers4"
	createdDate = "2021-05-06T14:49:37.871695Z"
)

// a map of fields in json vs db column names
var fieldMapper = map[string]string{
	"name":        "name",
	"id":          "id",
	"createdDate": "created_date",
}

// convertToSQLField : a mapper to whitelist the columns allowed in query
func convertToSQLField(key string) string {
	return fieldMapper[key]
}

// An entity to paginate
type Company struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CreatedDate string `json:"createdDate"`
}

// A slice which implements the pageable interface
type Companies []*Company

// UniqueVal : Returns the unique value to be used in cursor for pagination.
func (cc Companies) UniqueVal() *string {
	return &cc[len(cc)-1].ID
}

// SortingVal : Returns the default sorting value for this entity to be used in cursor for pagination.
func (cc Companies) SortingVal() *string {
	//default sorting val
	return &cc[len(cc)-1].Name
}

// DropLastRecord : Drops the last record from the slice
func (cc Companies) DropLastRecord() pagination.Pageable {
	cc = cc[:len(cc)-1]
	return cc
}

// LastRecord : Returns the last record from the slice
func (cc Companies) LastRecord() interface{} {
	return *cc[len(cc)-1]
}

// Length : Returns the length of slice
func (cc Companies) Length() int {
	return len(cc)
}

// getFilter : Gets pagination filter
func getFilter(u string) (*filter.Filter, error) {
	vals, _ := url.ParseQuery(u)
	limit, _ := strconv.Atoi(string(vals.Get("https://company-service/v1/companies?limit")))
	cursor := vals["cursor"]
	sortBy := vals["sortBy"]

	defaultSortBy := "name"
	companyByCompanyID := "id"

	start := time.Now()

	paginate := pagination.NewParams(limit, cursor[0], &sortBy[0])
	//handle pagination
	f, err := pagination.GetFilter(paginate, defaultSortBy, companyByCompanyID, nil,
		sql.GetConverter(), convertToSQLField)
	if err != nil {
		return nil, fmt.Errorf("GetByPartner | Failed to get pageFilter for params: %+v with error: %v", paginate, err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Time taken | getFilter : %s\n", elapsed)

	return f, nil
}

// paginateResults : will paginate the results returned and add a header for the next page link
func paginateResults(w http.ResponseWriter, r *http.Request, companies Companies, limit int, sortBy *string) (Companies, error) {

	start := time.Now()

	//append link in response header
	err := pagination.Paginate(w, *r, &companies, limit, sortBy)
	if err != nil {
		return nil, fmt.Errorf("failed to paginate companies: %+v | error: %v", companies, err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Time taken | paginateResults : %s\n", elapsed)

	return companies, nil

}

func main() {
	c := pagination.Cursor{
		UniqueID:    companyID,
		OrderingKey: createdDate,
	}

	//setting a mock cursor, limit and sortBy
	cursor, _ := c.Encode()
	limit := 2
	sortBy := "createdDate"

	expectedFilter := filter.New("(created_date,id) > (%v,%v) order by created_date limit %v", createdDate, companyID, strconv.Itoa(limit+1))

	u := `https://company-service/v1/companies?limit=%v&cursor=%v&sortBy=%v`

	ru := fmt.Sprintf(u, limit, cursor, sortBy)

	f, _ := getFilter(ru)

	if !(reflect.DeepEqual(f, expectedFilter)) {
		fmt.Printf("got: %v, want: %v", f, expectedFilter)
	}

	r, _ := http.NewRequest(http.MethodGet, u, nil)
	w := httptest.NewRecorder()

	//few dummy companies
	c1 := &Company{"1", "tes", "2021-07-07T17:14:53.906967433Z"}
	c2 := &Company{"2", "bor", "2021-07-07T17:14:53.906967433Z"}
	c3 := &Company{"3", "neuro", "2021-07-07T17:14:53.906967433Z"}

	dummyCompanies := Companies{c1, c2, c3}
	expectedCompanies := Companies{c1, c2}

	records, _ := paginateResults(w, r, dummyCompanies, 2, &sortBy)

	//expecting the first two records c1 and c2
	if !(reflect.DeepEqual(records, expectedCompanies)) {
		fmt.Printf("got: %v, want: %v", records, expectedCompanies)
	}

	//setting expectation for the next page link
	expectedCursor := pagination.Cursor{
		UniqueID:    c2.ID,
		OrderingKey: c2.CreatedDate,
	}
	expectedCursorEnc, _ := expectedCursor.Encode()
	expectedNextPageLink := fmt.Sprintf(u, limit, expectedCursorEnc, sortBy)

	//expecting next page link in the response header
	if !(reflect.DeepEqual(w.Header().Get("Link"), expectedNextPageLink)) {
		fmt.Printf("got: %v, want: %v", w.Header().Get("Link"), expectedNextPageLink)
	}
}
