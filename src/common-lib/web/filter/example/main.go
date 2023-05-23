package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/converters/sql"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web"
	restfilter "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/web/filter/strategies/tokenize"
)

const (
	getTicketsByPartnerIDSQL = `SELECT * FROM ticket WHERE partner_id = $1;`
	expectedSQL              = `SELECT * FROM ticket WHERE partner_id = $1 AND  ( summary  like  $2 or summary  like  $3 ) and ( (status) = ($4) or (status) = ($5) or partner_id in ($6,$7,$8,$9) );`
)

var fieldMapper = map[string]string{
	"summary":   "summary",
	"status":    "status",
	"partnerId": "partner_id",
}

// convertToSQLField : a mapper to whitelist the columns allowed in query
func convertToSQLField(key string) string {
	return fieldMapper[key]
}

// testHandler : A test handler to be wrapped with advanced filter handling.
func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Middleware: Request headers with query: %v\n", r.Header)

	partnerID := 123

	//Getting restfilter filter from request header.
	f, _ := restfilter.GetFilter(r)

	//Appending the putput to the where clause of the query.
	query, vals := sql.AppendFilterToWhereClause(getTicketsByPartnerIDSQL, f, []interface{}{partnerID})

	fmt.Println(query)
	fmt.Println(vals)

	//setting up a mock db.
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//setting query expectations.
	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).WithArgs(123, "%critical%", "%emergency%", "New", "InProgress", "100", "200", "300", "400").WillReturnRows(nil)

	//prepared statment call with the query and vals.
	db.Query(query, vals...)

	//asserting the expectation.
	if err = mock.ExpectationsWereMet(); err != nil {
		panic(fmt.Sprintf("There were unfulfilled expectations: %s", err))
	}
}

// restFilterErrorHandler : An error handler middleware that needs to be written to be able to handle the error returned by advanced filter handling middleware
func restFilterErrorHandler(r *http.Request, next http.HandlerFunc, hdlr restfilter.HandlerFunc, callback func(string) string) web.HTTPHandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var err error
		if req, err = hdlr(w, req, tokenize.GetStrategy(), sql.GetConverter(), callback); err != nil {
			fmt.Printf("Middleware: Error: %v\n", err.Error())
			return
		}
		next(w, req)
	}
}

// Advanced filter handling as a middlware
func asMiddleware(u string) {
	r, _ := http.NewRequest(http.MethodGet, u, nil)
	restFilterErrorHandler(r, testHandler, restfilter.Middleware, convertToSQLField)(nil, r)
}

// NoTree : Converts to SQL without a tree
func noMiddleware(u string) {
	res, _ := url.ParseQuery(u)
	filter := string(res.Get("https://ticket-service/v1/tickets?filter"))

	start := time.Now()
	ts := tokenize.GetStrategy()
	cnv := sql.GetConverter()
	out, err := ts.Parse(cnv, filter, convertToSQLField)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Filter syntax error: %v\nFilter: %v\n", err, filter)
	}
	fmt.Printf("Time taken: %s\n", elapsed)
	fmt.Printf("NoMiddleware: %v\n", out)
}

func main() {
	u := `https://ticket-service/v1/tickets?filter=(summary : "critical" OR summary : "emergency") AND (status = "New" OR status = "InProgress" OR partnerId IN 100,200,300,400)`
	noMiddleware(u)
	asMiddleware(u)
}
