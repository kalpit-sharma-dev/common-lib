package pagination

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/command"
	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/filter/converters/sql"
)

var (
	statusID = "00112233-4455-6677-8899-aabbccddeeff"

	limit       = 5
	uniqueField = "id"
	sortBy      = "name"
)

var fieldMap = map[string]string{
	"name":    "name",
	"id":      "id",
	"email":   "email",
	"address": "address",
}

var mapper = func(key string) string {
	v := fieldMap[key]
	if v == "" {
		fmt.Println(key)
		return "dummy"
	}
	return v
}

// A mock entity that's pageable
type Company struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// A slice of enties which is pageable
type Companies []*Company

func (cc Companies) UniqueVal() *string {
	return &cc[len(cc)-1].ID
}

func (cc Companies) SortingVal() *string {
	return &cc[len(cc)-1].Name
}

func TestGetFilter(t *testing.T) {
	type args struct {
		paginate           *PaginationParams
		defaultSortByField string
		uniqueIDField      string
		prefixFilter       *filter.Filter
		cnv                command.Converter
		fieldMapper        func(string) string
	}
	tests := []struct {
		name          string
		args          args
		want          *filter.Filter
		wantShouldAnd bool
		wantErr       bool
	}{
		{
			name: "success_with_prefix_with_cursor",
			args: args{
				paginate: func() *PaginationParams {
					cursor := Cursor{OrderingKey: companyName, UniqueID: companyID}
					encoded, _ := cursor.Encode()
					pp := NewParams(limit, encoded, &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       filter.New("activityStatus = %v", statusID),
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "activityStatus = %v AND (name,id) > (%v,%v) order by name, id limit %v"
				newLimit := limit + 1
				f := filter.New(query, statusID, companyName, companyID, strconv.Itoa(newLimit))
				return f
			}(),
			wantShouldAnd: true,
			wantErr:       false,
		},
		{
			name: "success_with_prefix_without_cursor",
			args: args{
				paginate: func() *PaginationParams {
					pp := NewParams(limit, "", &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       filter.New("activityStatus = %v", statusID),
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "activityStatus = %v order by name, id limit %v"
				newLimit := limit + 1
				f := filter.New(query, statusID, strconv.Itoa(newLimit))
				return f
			}(),
			wantShouldAnd: true,
			wantErr:       false,
		},
		{
			name: "success_with_prefix_with_cursor_and_sort_order",
			args: args{
				paginate: func() *PaginationParams {
					cursor := Cursor{OrderingKey: companyName, UniqueID: companyID}
					encoded, _ := cursor.Encode()
					sortBy := "name desc"
					pp := NewParams(limit, encoded, &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       filter.New("activityStatus = %v", statusID),
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "activityStatus = %v AND (name,id) < (%v,%v) order by name desc, id desc limit %v"
				newLimit := limit + 1
				f := filter.New(query, statusID, companyName, companyID, strconv.Itoa(newLimit))
				return f
			}(),
			wantShouldAnd: true,
			wantErr:       false,
		},
		{
			name: "success_with_sortBy_without_limit",
			args: args{
				paginate: func() *PaginationParams {
					pp := NewParams(0, "", &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       filter.New("activityStatus = %v", statusID),
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "activityStatus = %v order by name, id"
				f := filter.New(query, statusID)
				return f
			}(),
			wantShouldAnd: true,
			wantErr:       false,
		},
		{
			name: "success_with_sortBy_without_limit_and_prefix",
			args: args{
				paginate: func() *PaginationParams {
					pp := NewParams(0, "", &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       nil,
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "order by name, id"
				f := filter.New(query)
				f.ShouldAnd = false
				return f
			}(),
			wantShouldAnd: false,
			wantErr:       false,
		},
		{
			name: "success_without_prefix_with_cursor",
			args: args{
				paginate: func() *PaginationParams {
					cursor := Cursor{OrderingKey: companyName, UniqueID: companyID}
					encoded, _ := cursor.Encode()
					pp := NewParams(limit, encoded, &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       nil,
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "(name,id) > (%v,%v) order by name, id limit %v"
				newLimit := limit + 1
				f := filter.New(query, companyName, companyID, strconv.Itoa(newLimit))
				return f
			}(),
			wantShouldAnd: true,
			wantErr:       false,
		},
		{
			name: "success_without_prefix_without_cursor",
			args: args{
				paginate: func() *PaginationParams {
					pp := NewParams(limit, "", &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       nil,
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "order by name, id limit %v"
				newLimit := limit + 1
				f := filter.New(query, strconv.Itoa(newLimit))
				f.ShouldAnd = false
				return f
			}(),
			wantShouldAnd: false,
			wantErr:       false,
		},
		{
			name: "should_not_insert_uniqueField_if_already_in_sortBy",
			args: args{
				paginate: func() *PaginationParams {
					cursor := Cursor{OrderingKey: companyID, UniqueID: companyID}
					encoded, _ := cursor.Encode()
					pp := NewParams(limit, encoded, &uniqueField)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       nil,
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "(id) > (%v) order by id limit %v"
				newLimit := limit + 1
				f := filter.New(query, companyID, strconv.Itoa(newLimit))
				return f
			}(),
			wantShouldAnd: true,
			wantErr:       false,
		},
		{
			name: "should_create_filter_even_wih_excess_whitespace_in_sortBy",
			args: args{
				paginate: func() *PaginationParams {
					sortBy := "id, name , email,address   desc"
					pp := NewParams(limit, "", &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       nil,
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want: func() *filter.Filter {
				query := "order by id desc, name desc, email desc, address desc limit %v"
				newLimit := limit + 1
				f := filter.New(query, strconv.Itoa(newLimit))
				f.ShouldAnd = false
				return f
			}(),
			wantShouldAnd: false,
			wantErr:       false,
		},
		{
			name: "should_not_create_filter_for_empty_sortBy",
			args: args{
				paginate: func() *PaginationParams {
					sortBy := "  "
					pp := NewParams(limit, "", &sortBy)
					return pp
				}(),
				defaultSortByField: sortBy,
				uniqueIDField:      uniqueField,
				prefixFilter:       nil,
				cnv:                sql.GetConverter(),
				fieldMapper:        mapper,
			},
			want:          nil,
			wantShouldAnd: false,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFilter(tt.args.paginate, tt.args.defaultSortByField, tt.args.uniqueIDField, tt.args.prefixFilter, tt.args.cnv, tt.args.fieldMapper)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFilter() = %v, want %v", got, tt.want)
			}
			if got != nil && (got.ShouldAnd != tt.wantShouldAnd) {
				t.Errorf("ShouldAnd = %v, want %v", got.ShouldAnd, tt.wantShouldAnd)
			}
		})
	}
}

func TestPaginate(t *testing.T) {

	c1 := &Company{"1", "ABC Corp"}
	c2 := &Company{"2", "DEF Corp"}
	c3 := &Company{"3", "GHI Corp"}

	type args struct {
		w            http.ResponseWriter
		r            http.Request
		records      Companies
		limit        int
		customSortBy *string
	}
	tests := []struct {
		name     string
		args     args
		want     Companies
		wantLink string
		wantErr  bool
	}{
		{
			name: "success",
			args: args{
				w: &httptest.ResponseRecorder{},
				r: func() http.Request {
					url := "https://somewhere.com/companies?limit=2&sortBy=name"
					r, _ := http.NewRequest(http.MethodGet, url, nil)
					return *r
				}(),
				records: Companies{c1, c2, c3},
				limit:   2,
				customSortBy: func() *string {
					s := "name"
					return &s
				}(),
			},
			want: Companies{c1, c2},
			wantLink: func() string {
				myURL := `</companies?cursor=%v&limit=2&sortBy=name>; rel="next"`
				c := Cursor{
					UniqueID:    c2.ID,
					OrderingKey: c2.Name,
				}
				encoded, _ := c.Encode()
				myURL = fmt.Sprintf(myURL, url.QueryEscape(encoded))
				return myURL
			}(),
			wantErr: false,
		},
		{
			name: "success_with_sort_order",
			args: args{
				w: &httptest.ResponseRecorder{},
				r: func() http.Request {
					url := "https://somewhere.com/companies?limit=2&sortBy=name+desc"
					r, _ := http.NewRequest(http.MethodGet, url, nil)
					return *r
				}(),
				records: Companies{c1, c2, c3},
				limit:   2,
				customSortBy: func() *string {
					s := "name desc"
					return &s
				}(),
			},
			want: Companies{c1, c2},
			wantLink: func() string {
				myURL := `</companies?cursor=%v&limit=2&sortBy=name+desc>; rel="next"`
				c := Cursor{
					UniqueID:    c2.ID,
					OrderingKey: c2.Name,
				}
				encoded, _ := c.Encode()
				myURL = fmt.Sprintf(myURL, url.QueryEscape(encoded))
				return myURL
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Paginate(tt.args.w, tt.args.r, &tt.args.records, tt.args.limit, tt.args.customSortBy)
			if (err != nil) != tt.wantErr {
				t.Errorf("Paginate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.args.records, tt.want) {
				t.Errorf("Paginate() = %v, want %v", tt.args.records, tt.want)
			}
			if !reflect.DeepEqual(tt.args.w.Header().Get(nextLinkHeader), tt.wantLink) {
				t.Errorf("Paginate() = %v, want %v", tt.args.w.Header().Get(nextLinkHeader), tt.wantLink)
			}
		})
	}
}

func TestTransformSortBy(t *testing.T) {
	var testCases = []struct {
		name     string
		sortBy   *string
		expected *string
	}{
		{
			name:     "no_sort_if_nil",
			sortBy:   nil,
			expected: nil,
		},
		{
			name:     "no_sort_if_empty_string",
			sortBy:   strPtr(""),
			expected: nil,
		},
		{
			name:     "no_sort_if_whitespace_only_string",
			sortBy:   strPtr("    "),
			expected: nil,
		},
		{
			name:     "empty_sorts_if_just_a_comma",
			sortBy:   strPtr(","),
			expected: strPtr(","),
		},
		{
			name:     "success_with_one_field",
			sortBy:   strPtr("id"),
			expected: strPtr("id"),
		},
		{
			name:     "success_with_one_field_and_order",
			sortBy:   strPtr("id ASC"),
			expected: strPtr("id asc"),
		},
		{
			name:     "success_with_no_spaces",
			sortBy:   strPtr("id,name"),
			expected: strPtr("id,name"),
		},
		{
			name:     "success_with_space_after_comma",
			sortBy:   strPtr("id,name, username"),
			expected: strPtr("id,name,username"),
		},
		{
			name:     "succes_with_space_around_comma",
			sortBy:   strPtr("id,name, username , password"),
			expected: strPtr("id,name,username,password"),
		},
		{
			name:     "success_with_uppercase_sortOrder",
			sortBy:   strPtr("id,name, username , password DESC"),
			expected: strPtr("id,name,username,password desc"),
		},
		{
			name:     "success_with_extra_space_before_sortOrder",
			sortBy:   strPtr("id,name, username , password    desc"),
			expected: strPtr("id,name,username,password desc"),
		},
		{
			name:     "drop_field_after_space_if_not_valid_sortOrder",
			sortBy:   strPtr("id,name, username , password    desccc"),
			expected: strPtr("id,name,username,password"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := transformSortBy(testCase.sortBy)
			if testCase.expected != nil && actual != nil {
				// This provides better error messages when debugging tests
				assert.Equal(t, *testCase.expected, *actual)
			} else {
				assert.Equal(t, testCase.expected, actual)
			}
		})
	}
}

func strPtr(str string) *string {
	return &str
}
