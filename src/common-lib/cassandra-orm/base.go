package cassandraorm

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/helpers"
)

type (
	// Base interface
	Base interface {
		BaseWithBatch

		// Keys get table keys
		Keys() []string
		// Table get table name
		Table() string
		// GetColumns get table columns
		GetColumns() []string
		// Quote use in case of camel case DB field names
		Quote(s string) string

		// All get all rows
		All(values interface{}, keyCols ...interface{}) error
		// Get get one row
		Get(value interface{}, keyCols ...interface{}) error
		// Add checks/generates ID and inserts item
		Add(item Model) error
		// AddWithTTL checks/generates ID and inserts item with ttl
		AddWithTTL(item Model, ttl time.Duration) error
		// Update item in repository
		Update(item Model) error
		// UpdateWithTTL item in repository with ttl
		UpdateWithTTL(item Model, ttl time.Duration) error
		// Delete item
		Delete(item Model) error

		// ExecuteBatch execute batch
		ExecuteBatch(batch *gocql.Batch) error
		// ExecRelease performs exec and releases the query, a released query cannot be reused
		ExecRelease(q *gocqlx.Queryx) error
		// QuerySelect is a convenience function for creating iterator and calling Select
		QuerySelect(values interface{}, queryBuilder qb.Builder, params map[string]interface{}) error
		// QuerySelectPagination is a convenience function for creating iterator and calling Select with pagination
		QuerySelectPagination(values interface{}, queryBuilder qb.Builder, params map[string]interface{}, pageSize, page int) error
		// QueryGet is a convenience function for creating iterator and calling Get
		QueryGet(value interface{}, queryBuilder qb.Builder, params map[string]interface{}) error

		// Register set observer for repo
		Register(observer Observer)
		// Deregister revert observer to default (empty)
		Deregister()
	}

	baseStrategy interface {
		// Add check/generates ID and inserts item
		Add(b *base, item Model) error
		// AddWithTTL checks/generates ID and inserts item with ttl
		AddWithTTL(b *base, item Model, ttl time.Duration) error
		// Update item in repository
		Update(b *base, item Model) error
		// UpdateWithTTL item in repository
		UpdateWithTTL(b *base, item Model, ttl time.Duration) error
		// Delete delete item
		Delete(b *base, item Model) error
	}

	base struct {
		item                Model
		table               string
		keys                []string
		columns             []string
		columnsName         []string
		columnsNameWithTags []string
		viewTables          map[string][]string
		execRelease         func(q *gocqlx.Queryx) error
		observer            Observer
		strategy            baseStrategy
	}

	simpleBase struct{}

	baseWithBatch struct{}
)

// NewBase creates new instance
func NewBase(item Model, tableName string, tableKeys []string, viewTables map[string][]string) Base {
	instance := &base{
		item:       item,
		table:      tableName,
		keys:       tableKeys,
		observer:   &DefaultObserver{},
		viewTables: viewTables,
		strategy:   getStrategy(viewTables),
	}
	instance.init()
	return instance
}

func getStrategy(tables map[string][]string) baseStrategy {
	if len(tables) == 0 {
		return &simpleBase{}
	}
	return &baseWithBatch{}
}

func (b *base) Keys() []string {
	return b.keys
}

func (b *base) Table() string {
	return b.table
}

func (b *base) GetColumns() []string {
	return b.columnsName
}

func (b *base) Quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func (b *base) Get(value interface{}, keyCols ...interface{}) error {
	return b.getFromTable(b.Table(), b.Keys(), value, keyCols...)
}

func (b *base) All(values interface{}, keyCols ...interface{}) error {
	return b.allFromTable(b.Table(), b.Keys(), values, keyCols...)
}

func (b *base) init() {
	b.execRelease = func(q *gocqlx.Queryx) error {
		return q.ExecRelease()
	}
	b.columns, b.columnsName, b.columnsNameWithTags = getColumnsAndColumnNames(b.item, b)
}

func (b *base) Delete(item Model) error {
	return b.exec(item, b.strategy.Delete, EventBeforeDelete, EventAfterDelete)
}

func (b *base) ExecuteBatch(batch *gocql.Batch) error {
	return Session.ExecuteBatch(batch)
}
func (b *base) Update(item Model) error {
	return b.exec(item, b.strategy.Update, EventBeforeUpdate, EventAfterUpdate)
}

func (b *base) UpdateWithTTL(item Model, ttl time.Duration) error {
	updateWithTTL := func(b *base, item Model) error {
		return b.strategy.UpdateWithTTL(b, item, ttl)
	}
	return b.exec(item, updateWithTTL, EventBeforeUpdate, EventAfterUpdate)
}

func (b *base) Add(item Model) error {
	return b.exec(item, b.strategy.Add, EventBeforeAdd, EventAfterAdd)
}

func (b *base) AddWithTTL(item Model, ttl time.Duration) error {
	addWithTTL := func(b *base, item Model) error {
		return b.strategy.AddWithTTL(b, item, ttl)
	}
	return b.exec(item, addWithTTL, EventBeforeAdd, EventAfterAdd)
}

func batchFailed(err error) bool {
	return err != nil && strings.Contains(err.Error(), BatchTooLarge)
}

func (b *base) execBatchSeparately(batchFunc func(*gocql.Batch, Model) error, item Model) error {
	batch := Session.NewBatch(gocql.LoggedBatch)
	err := batchFunc(batch, item)
	if err != nil {
		return err
	}
	for _, entry := range batch.Entries {
		query := gocqlx.Query(Session.Query(entry.Stmt), b.columnsNameWithTags).Bind(entry.Args...)
		err = query.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *base) exec(item Model, execFunc func(*base, Model) error, beforeEvent, afterEvent EventType) error {
	b.observer.OnNotify(beforeEvent, item)
	err := execFunc(b, item)
	if err != nil {
		return err
	}
	b.observer.OnNotify(afterEvent, item)
	return nil
}

func (*simpleBase) Delete(b *base, item Model) error {
	keyCols, err := GetQueryKeys(item, b.Keys())
	if err != nil {
		return err
	}
	cmp, args := b.getBaseComparatorsAndArgs(keyCols...)
	stmt, names := qb.Delete(b.Table()).Where(cmp...).ToCql()
	query := gocqlx.Query(Session.Query(stmt), names).BindMap(args)
	return b.ExecRelease(query)
}

func (*simpleBase) Update(b *base, item Model) (err error) {
	return b.insert(item)
}

func (*simpleBase) UpdateWithTTL(b *base, item Model, ttl time.Duration) (err error) {
	return b.insertWithTTL(item, ttl)
}

func (*simpleBase) Add(b *base, item Model) (err error) {
	if err = item.AcquireID(); err != nil {
		return err
	}
	return b.insert(item)
}

func (*simpleBase) AddWithTTL(b *base, item Model, ttl time.Duration) (err error) {
	if err = item.AcquireID(); err != nil {
		return err
	}
	return b.insertWithTTL(item, ttl)
}

func (*baseWithBatch) Delete(b *base, item Model) error {
	return batchExec(b, b.DeleteWithBatch, item)
}

func (*baseWithBatch) Update(b *base, item Model) (err error) {
	return batchExec(b, b.UpdateWithBatch, item)
}

func (*baseWithBatch) UpdateWithTTL(b *base, item Model, ttl time.Duration) (err error) {
	updateWithTTL := func(batch *gocql.Batch, item Model) error {
		return b.UpdateWithBatchAndTTL(batch, item, ttl)
	}
	return batchExec(b, updateWithTTL, item)
}

func (*baseWithBatch) Add(b *base, item Model) (err error) {
	return batchExec(b, b.AddWithBatch, item)
}

func (*baseWithBatch) AddWithTTL(b *base, item Model, ttl time.Duration) (err error) {
	addWithTTL := func(batch *gocql.Batch, item Model) error {
		return b.AddWithBatchAndTTL(batch, item, ttl)
	}
	return batchExec(b, addWithTTL, item)
}

func batchExec(b *base, batchFunc func(*gocql.Batch, Model) error, item Model) error {
	batch := Session.NewBatch(gocql.LoggedBatch)
	err := batchFunc(batch, item)
	if err != nil {
		return err
	}
	err = Session.ExecuteBatch(batch)
	if err != nil {
		if !batchFailed(err) {
			return err
		}
		return b.execBatchSeparately(batchFunc, item)
	}
	return nil
}

func (b *base) ExecRelease(q *gocqlx.Queryx) error {
	return b.execRelease(q)
}

func (b *base) QuerySelect(values interface{}, queryBuilder qb.Builder, params map[string]interface{}) error {
	stmt, names := queryBuilder.ToCql()
	query := gocqlx.Query(Session.Query(stmt), names).BindMap(params)
	defer query.Release()
	err := gocqlx.Select(values, query.Query)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(values)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Len() == 0 {
		return gocql.ErrNotFound
	}
	return nil
}

func (b *base) QuerySelectPagination(values interface{}, queryBuilder qb.Builder, params map[string]interface{}, pageSize, page int) error {
	stmt, names := queryBuilder.ToCql()
	q := Session.Query(stmt)
	query := gocqlx.Query(q, names).BindMap(params)
	defer query.Release()

	query.PageState(nil).PageSize(pageSize)
	iter := query.Iter()

	for i := 1; i < page; i++ {
		if len(iter.PageState()) > 0 {
			// move iterator to next page
			iter = query.PageState(iter.PageState()).Iter()
		}
	}

	err := iter.Close()
	if err != nil {
		return err
	}

	err = gocqlx.Select(values, query.Query)
	if err != nil {
		return err
	}

	val := reflect.ValueOf(values)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Len() == 0 {
		return gocql.ErrNotFound
	}
	return nil
}

func (b *base) QueryGet(value interface{}, queryBuilder qb.Builder, params map[string]interface{}) error {
	stmt, names := queryBuilder.ToCql()
	query := gocqlx.Query(Session.Query(stmt), names).BindMap(params)
	defer query.Release()
	return gocqlx.Get(value, query.Query)
}

func (b *base) Register(observer Observer) {
	b.observer = observer
}

func (b *base) Deregister() {
	b.observer = &DefaultObserver{}
}

// getFromTable get one row from given table
func (b *base) getFromTable(table string, keys []string, value interface{}, keyCols ...interface{}) error {
	cmp, args := b.getComparatorsAndArgs(keys, keyCols...)
	queryBuilder := qb.Select(table).Where(cmp...).Limit(1)
	return b.QueryGet(value, queryBuilder, args)
}

// allFromTable get all rows from given table
func (b *base) allFromTable(table string, keys []string, values interface{}, keyCols ...interface{}) error {
	cmp, args := b.getComparatorsAndArgs(keys, keyCols...)
	queryBuilder := qb.Select(table).Where(cmp...)
	return b.QuerySelect(values, queryBuilder, args)
}

func (b *base) insert(item Model) error {
	_, query := b.getInsertStmtAndQuery(item)
	return b.ExecRelease(query)
}

func (b *base) insertWithTTL(item Model, ttl time.Duration) error {
	row, err := helpers.SerializeDefault(item)
	if err != nil {
		return err
	}
	_, query := b.getInsertStmtAndQueryWithTTL(b.Table(), row, ttl)
	return b.ExecRelease(query)
}

func (b *base) getInsertStmtAndQuery(item Model) (string, *gocqlx.Queryx) {
	stmt, _ := qb.Insert(b.Table()).Columns(b.columns...).ToCql()
	query := gocqlx.Query(Session.Query(stmt), b.columnsNameWithTags).BindStruct(item)
	return stmt, query
}

func (b *base) getInsertStmtAndQueryWithTTL(tableName string, row map[string]interface{}, ttl time.Duration) (string, *gocqlx.Queryx) {
	columns := b.GetColumns()
	fields := make([]string, 0, len(columns))
	values := make(map[string]interface{}, len(columns)+1)
	for _, column := range columns {
		fields = append(fields, b.Quote(column))
		values[b.Quote(column)] = row[column]
	}

	values["_ttl"] = int(ttl.Seconds())
	stmt, names := qb.Insert(tableName).Columns(fields...).TTL().ToCql()
	query := gocqlx.Query(Session.Query(stmt), names).BindMap(values)

	return stmt, query
}

func (b *base) getBaseComparatorsAndArgs(keyCols ...interface{}) ([]qb.Cmp, map[string]interface{}) {
	return b.getComparatorsAndArgs(b.keys, keyCols...)
}

func (b *base) getComparatorsAndArgs(keys []string, keyCols ...interface{}) ([]qb.Cmp, map[string]interface{}) {
	comparators := make([]qb.Cmp, len(keyCols))
	args := make(map[string]interface{}, len(keyCols))

	for i, val := range keyCols {
		comparators[i] = qb.Eq(b.Quote(keys[i]))
		args[b.Quote(keys[i])] = val
	}

	return comparators, args
}

func getColumnsAndColumnNames(item Model, b Base) (columns, columnsName, columnsNameWithTags []string) {
	if item == nil {
		return
	}
	t := reflect.TypeOf(item).Elem()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("db")
		if tag == "-" {
			continue
		}

		column := b.Quote(t.Field(i).Name)
		columnName := t.Field(i).Name
		if strings.TrimSpace(tag) != "" {
			column = b.Quote(tag)
			columnName = tag
		}
		columns = append(columns, column)
		columnsName = append(columnsName, t.Field(i).Name)
		columnsNameWithTags = append(columnsNameWithTags, columnName)
	}
	return
}
