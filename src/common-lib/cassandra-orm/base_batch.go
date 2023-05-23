package cassandraorm

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/qb"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm/helpers"
)

// BatchTooLarge identifies batches that exceed size limit and can't be executed
const BatchTooLarge = "Batch too large"

// BaseWithBatch interface
type BaseWithBatch interface {
	// AllFromTable get all rows from given table
	AllFromTable(table string, values interface{}, keyCols ...interface{}) error
	// GetFromTable get one row from given table
	GetFromTable(table string, value interface{}, keyCols ...interface{}) error

	// AddWithBatch checks/generates ID and inserts item (consumer is responsible for executing batch)
	AddWithBatch(batch *gocql.Batch, item Model) error
	// AddWithBatchAndTTL checks/generates ID and inserts item with ttl (consumer is responsible for executing batch)
	AddWithBatchAndTTL(batch *gocql.Batch, item Model, ttl time.Duration) error
	// UpdateWithBatch update item in repository (consumer is responsible for executing batch)
	UpdateWithBatch(batch *gocql.Batch, item Model) error
	// UpdateWithBatchAndTTL update item in repository with ttl(consumer is responsible for executing batch)
	UpdateWithBatchAndTTL(batch *gocql.Batch, item Model, ttl time.Duration) error
	// DeleteWithBatch delete item (consumer is responsible for executing batch)
	DeleteWithBatch(batch *gocql.Batch, item Model) error
}

func (b *base) GetFromTable(table string, value interface{}, keyCols ...interface{}) error {
	return b.getFromTable(table, b.viewTables[table], value, keyCols...)
}

func (b *base) AllFromTable(table string, values interface{}, keyCols ...interface{}) error {
	return b.allFromTable(table, b.viewTables[table], values, keyCols...)
}

func (b *base) DeleteWithBatch(batch *gocql.Batch, item Model) error {
	keyCols, err := GetQueryKeys(item, b.Keys())
	if err != nil {
		return err
	}
	cmp, _ := b.getBaseComparatorsAndArgs(keyCols...)
	stmt, _ := qb.Delete(b.Table()).Where(cmp...).ToCql()
	batch.Query(stmt, keyCols...)
	return b.addViewsDeleteBatch(item, batch)
}

func (b *base) UpdateWithBatch(batch *gocql.Batch, item Model) error {
	return b.insertWithBatch(batch, item)
}

func (b *base) UpdateWithBatchAndTTL(batch *gocql.Batch, item Model, ttl time.Duration) error {
	return b.insertWithBatchAndTTL(batch, item, ttl)
}

func (b *base) AddWithBatch(batch *gocql.Batch, item Model) error {
	if err := item.AcquireID(); err != nil {
		return err
	}
	return b.insertWithBatch(batch, item)
}

func (b *base) AddWithBatchAndTTL(batch *gocql.Batch, item Model, ttl time.Duration) error {
	if err := item.AcquireID(); err != nil {
		return err
	}
	return b.insertWithBatchAndTTL(batch, item, ttl)
}

func (b *base) insertWithBatch(batch *gocql.Batch, item Model) error {
	newRow, err := helpers.SerializeDefault(item)
	if err != nil {
		return err
	}
	stmt, _ := b.getInsertStmtAndQuery(item)
	m := make([]interface{}, len(b.columnsName))
	for i, name := range b.columnsName {
		m[i] = newRow[name]
	}
	batch.Query(stmt, m...)

	err = b.deleteOld(newRow, batch)
	if err != nil {
		return err
	}

	for tableName := range b.viewTables {
		stmt, _ := qb.Insert(tableName).Columns(b.columns...).ToCql()
		batch.Query(stmt, m...)
	}
	return nil
}

func (b *base) insertWithBatchAndTTL(batch *gocql.Batch, item Model, ttl time.Duration) error {
	newRow, err := helpers.SerializeDefault(item)
	if err != nil {
		return err
	}
	stmt, _ := b.getInsertStmtAndQueryWithTTL(b.Table(), newRow, ttl)
	m := make([]interface{}, len(b.columnsName), len(b.columnsName)+1)
	for i, name := range b.columnsName {
		m[i] = newRow[name]
	}
	m = append(m, int(ttl.Seconds()))
	batch.Query(stmt, m...)

	err = b.deleteOld(newRow, batch)
	if err != nil {
		return err
	}

	for tableName := range b.viewTables {
		stmt, _ := b.getInsertStmtAndQueryWithTTL(tableName, newRow, ttl)
		batch.Query(stmt, m...)
	}
	return nil
}

func (b *base) deleteOld(newRow map[string]interface{}, batch *gocql.Batch) error {
	old, err := b.getOld(newRow)
	if err != nil {
		if err != gocql.ErrNotFound {
			return err
		}
		return nil
	}

	oldRow, err := helpers.SerializeDefault(old)
	if err != nil {
		return err
	}
	for tableName, tableKeys := range b.viewTables {
		if !isRowsEqualByKeys(oldRow, newRow, tableKeys) {
			if err = b.fillDeleteBatch(old, tableName, tableKeys, batch); err != nil {
				return err
			}
		}
	}
	return nil
}

func isRowsEqualByKeys(first, second map[string]interface{}, keys []string) bool {
	for _, key := range keys {
		if first[key] != second[key] {
			return false
		}
	}
	return true
}

func (b *base) addViewsDeleteBatch(item Model, batch *gocql.Batch) error {
	itemRow, err := helpers.SerializeDefault(item)
	if err != nil {
		return err
	}
	old, err := b.getOld(itemRow)
	if err != nil {
		if err == gocql.ErrNotFound {
			return nil
		}
		return err
	}

	for tableName, tableKeys := range b.viewTables {
		if err := b.fillDeleteBatch(old, tableName, tableKeys, batch); err != nil {
			return err
		}
	}
	return nil
}

func (b *base) fillDeleteBatch(item Model, tableName string, tableKeys []string, batch *gocql.Batch) error {
	keys, err := GetQueryKeys(item, tableKeys)
	if err != nil {
		return err
	}
	cmp, _ := b.getComparatorsAndArgs(tableKeys, keys...)
	stmt, _ := qb.Delete(tableName).Where(cmp...).ToCql()
	batch.Query(stmt, keys...)
	return nil
}

func (b *base) getOld(itemRows map[string]interface{}) (Model, error) {
	byKeys := make([]interface{}, len(b.Keys()))
	for i, name := range b.Keys() {
		byKeys[i] = itemRows[name]
	}

	old := b.item
	if err := b.Get(old, byKeys...); err != nil {
		return nil, err
	}
	return old, nil
}
