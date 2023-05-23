package main

import (
	"database/sql"
	"fmt"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/command"
	sqlcnv "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/converters/sql"
)

const (
	siteByPrimaryFlag = "primaryFlag"
	siteByActiveFlag  = "activeFlag"

	partnerID = "1234"
)

const (
	//a sql query used by a repo method
	getSitesByPartnerIDSQL = `SELECT id, name FROM sites WHERE partner_id = $1;`

	//expected query post appending the filter the to static query above
	expectedSQL = `SELECT id, name FROM sites WHERE partner_id = $1 AND primary_flag = $2 and active_flag = $3;`
)

// a dummy site entity
type site struct{}

// holds a map of whitelisted fields to db field names
var fieldMap map[string]string = map[string]string{
	"id":          "id",
	"name":        "name",
	"primaryFlag": "primary_flag",
	"activeFlag":  "active_flag",
}

// a repository interface to get sites
type SiteRepo interface {
	//will return a list of sites for the partner id, filtered by the filter f
	GetSites(partnerId string, f *filter.Filter) ([]site, error)

	//a method that will accept a field name and return the corresponding db column name
	FieldMapper() func(string) string
}

// a SQL specific implementation of repository
type SQLSiteRepo struct {
	db *sql.DB
}

// FieldMapper : returns the field mapper for the repository
func (r *SQLSiteRepo) FieldMapper() func(string) string {
	return func(key string) string {
		return fieldMap[key]
	}
}

// GetSites : returns a list of sites by partnerId, filtered by the filter f
func (r *SQLSiteRepo) GetSites(partnerID string, f *filter.Filter) ([]site, error) {

	//Appending the putput to the where clause of the query.
	query, vals := sqlcnv.AppendFilterToWhereClause(getSitesByPartnerIDSQL, f, []interface{}{partnerID})

	//print and see the query and vals
	fmt.Println(query)
	fmt.Println(vals)

	//prepared statment call with the query and vals.
	_, err := r.db.Query(query, vals...)

	return []site{}, err
}

func main() {

	//setting up a mock db.
	db, mock, err := sqlmock.New()
	if err != nil {
		fmt.Printf("\nError creating mock db: %v", err)
	}
	defer db.Close()

	//setting query expectations.
	mock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).WithArgs(partnerID, true, true).WillReturnRows(nil)

	//Create the repo
	siteRepo := &SQLSiteRepo{db}

	//Get the desired converter, SQL in this case
	cnv := sqlcnv.GetConverter()

	//Get only the sites which are primary and active
	//Filter site records by company id
	primaryCmd := command.New(siteByPrimaryFlag, string(command.Eq), "")
	primaryFilter, err := cnv.DoForCommandWithValue(primaryCmd, siteRepo.FieldMapper())
	if err != nil {
		fmt.Printf("\nCommand: %+v | Failed to create filter with error:%v",
			primaryCmd, err)
		return
	}

	//Filter site records by company id
	activeCmd := command.New(siteByActiveFlag, string(command.Eq), "")
	activeFilter, err := cnv.DoForCommandWithValue(activeCmd, siteRepo.FieldMapper())
	if err != nil {
		fmt.Printf("\nCommand: %+v | Failed to create filter with error:%v",
			activeCmd, err)
		return
	}

	f := cnv.AND(primaryFilter.CopyWithNewVals(true), activeFilter.CopyWithNewVals(true))

	siteRepo.GetSites(partnerID, f)

	//asserting the expectation.
	if err = mock.ExpectationsWereMet(); err != nil {
		fmt.Printf("\nThere were unfulfilled expectations: %v", err)
	}
}
