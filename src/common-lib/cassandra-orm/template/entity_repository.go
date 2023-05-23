package template

import (
	"sync"
	"time"

	"github.com/cheekybits/genny/generic"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/qb"
	db "gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/cassandra-orm"
)

const entityBaseTableName = "entities"

var (
	entityKeyColumns = []string{"ID"}
	entityViewTables = map[string][]string{}

	entitiesRepository *EntityRepository
	initEntities       sync.Once
)

// EntityRepository interface
type EntityRepository struct {
	base db.Base
}

// IDType used for specifying entity type
type IDType generic.Type

// Entities singleton, thread-safe, returns pointer to Entities repository
func Entities() *EntityRepository {
	initEntities.Do(func() {
		entitiesRepository = &EntityRepository{base: db.NewBase(
			&Entity{},
			entityBaseTableName,
			entityKeyColumns,
			entityViewTables,
		)}
	})
	return entitiesRepository
}

// Get return item
func (r *EntityRepository) Get(keyCols ...interface{}) (*Entity, error) {
	entity := new(Entity)
	if err := r.base.Get(entity, keyCols...); err != nil {
		return nil, err
	}
	return entity, nil
}

// All returns slice of Entities
func (r *EntityRepository) All() ([]*Entity, error) {
	var entities []*Entity
	if err := r.base.All(&entities); err != nil {
		return nil, err
	}
	return entities, nil
}

// Add check/generates ID and inserts item
// nolint:interfacer
func (r *EntityRepository) Add(item *Entity) error {
	return r.base.Add(item)
}

// AddWithTTL check/generates ID and inserts item with ttl
// nolint:interfacer
func (r *EntityRepository) AddWithTTL(item *Entity, ttl time.Duration) error {
	return r.base.AddWithTTL(item, ttl)
}

// Update item in repository
// nolint:interfacer
func (r *EntityRepository) Update(item *Entity) error {
	return r.base.Update(item)
}

// UpdateWithTTL item in repository with ttl
// nolint:interfacer
func (r *EntityRepository) UpdateWithTTL(item *Entity, ttl time.Duration) error {
	return r.base.UpdateWithTTL(item, ttl)
}

// Delete item from repository
// nolint:interfacer
func (r *EntityRepository) Delete(item *Entity) error {
	return r.base.Delete(item)
}

// AddWithBatch adds all queries to batch
// nolint:interfacer
func (r *EntityRepository) AddWithBatch(batch *gocql.Batch, item *Entity) error {
	return r.base.AddWithBatch(batch, item)
}

// AddWithBatchAndTTL adds all queries with ttl to batch
// nolint:interfacer
func (r *EntityRepository) AddWithBatchAndTTL(batch *gocql.Batch, item *Entity, ttl time.Duration) error {
	return r.base.AddWithBatchAndTTL(batch, item, ttl)
}

// UpdateWithBatch adds all queries to batch
// nolint:interfacer
func (r *EntityRepository) UpdateWithBatch(batch *gocql.Batch, item *Entity) error {
	return r.base.UpdateWithBatch(batch, item)
}

// UpdateWithBatchAndTTL adds all queries with ttl to batch
// nolint:interfacer
func (r *EntityRepository) UpdateWithBatchAndTTL(batch *gocql.Batch, item *Entity, ttl time.Duration) error {
	return r.base.UpdateWithBatchAndTTL(batch, item, ttl)
}

// DeleteWithBatch adds all queries to batch
// nolint:interfacer
func (r *EntityRepository) DeleteWithBatch(batch *gocql.Batch, item *Entity) error {
	return r.base.DeleteWithBatch(batch, item)
}

// GetByID returns item by @id or error if item doesn't exists in repository
func (r *EntityRepository) GetByID(id IDType) (*Entity, error) {
	return r.Get(id)
}

// GetByIDs returns slice of entities by slice of ids
func (r *EntityRepository) GetByIDs(ids ...IDType) ([]*Entity, error) {
	var entities []*Entity
	queryBuilder := qb.Select(r.base.Table()).Where(
		qb.In(r.base.Quote(r.base.Keys()[0])),
	)

	err := r.base.QuerySelect(&entities, queryBuilder, map[string]interface{}{
		r.base.Quote(r.base.Keys()[0]): ids,
	})
	return entities, err
}

// GetRows returns slice of Entities filtered by received keys
func (r *EntityRepository) GetRows(keyCols ...interface{}) ([]*Entity, error) {
	var entities []*Entity
	if err := r.base.All(&entities, keyCols...); err != nil {
		return nil, err
	}
	return entities, nil
}
