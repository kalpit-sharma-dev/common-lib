package sqlite

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // To load sqlite drivers
)

//mutex sync.Mutex

const dialect = "sqlite3" //As per ORM specification database type is named as a dialect;

type sqlite struct {
	connection *gorm.DB
	config     *Config
	mutex      sync.Mutex
}

//GetService is a function to return service instance
func GetService(config *Config) Service {
	return &sqlite{config: config}
}

func (s *sqlite) Init() error {
	var err error
	if s.connection != nil {
		err = s.connection.Close()
		if err != nil {
			return fmt.Errorf("failed to close database connection for config : %+v with error : %+v", s.config, err)
		}
	}

	if s.config == nil || s.config.DBName == "" {
		return fmt.Errorf("database connection to config : %+v is not available", s.config)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.connection, err = gorm.Open(dialect, s.config.DBName)
	if err != nil || s.connection == nil {
		return fmt.Errorf("failed to initilaize database connection for config : %+v with error : %+v", s.config, err)
	}

	s.connection.Exec("PRAGMA auto_vacuum = INCREMENTAL;")
	return nil
}

func (s *sqlite) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.connection != nil {
		err := s.connection.Close()
		if err != nil {
			return fmt.Errorf("failed to close connection : %+v", err)
		}
		s.connection = nil
	}
	return nil
}

//CreateTable is a function to Create a table in case this is not exist and update if this exists
func (s *sqlite) CreateTable(table interface{}) error {
	if s.connection == nil || table == nil {
		return fmt.Errorf("create Table :: Database connection to config : %+v is not available", s.config)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	tx := s.connection.Begin()

	if tx.Error != nil {
		return fmt.Errorf("failed to initialize transaction : %+v", tx.Error)
	}

	if !s.connection.HasTable(table) {
		tx.CreateTable(table)
	} else {
		tx.AutoMigrate(table)
	}
	return s.commit(tx)
}

func (s *sqlite) Add(record interface{}) error {
	if s.connection != nil {
		return s.connection.Create(record).Error
	}
	return fmt.Errorf("add :: Database connection is not available")
}

func (s *sqlite) AddAll(records []interface{}) error {
	tx := s.connection.Begin()

	if tx.Error != nil {
		return fmt.Errorf("failed to initialize transaction : %+v", tx.Error)
	}
	for _, record := range records {
		tx.Create(record)
		if tx.Error != nil {
			return s.commit(tx)
		}
	}
	return s.commit(tx)
}

//It will update all fields, even it is not changed
func (s *sqlite) Update(record interface{}) error {
	return s.connection.Save(record).Error
}

func (s *sqlite) Delete(record interface{}) error {
	return s.connection.Delete(record).Error

}

func (s *sqlite) DeleteWhere(whereQuery, whereArgs, record interface{}) error {
	return s.connection.Where(whereQuery, whereArgs).Delete(record).Error
}

func (s *sqlite) FirstOrCreate(where, out interface{}) error {
	return s.connection.Where(where).FirstOrCreate(out).Error
}

func (s *sqlite) First(out interface{}, where ...interface{}) error {
	if s.connection != nil {
		s.connection.First(out, where...)
	}
	return nil
}

func (s *sqlite) Count(table string) int {
	var recordCount int
	if s.connection != nil {
		s.connection.Table(table).Count(&recordCount)
	}
	return recordCount
}

func (s *sqlite) Get(limit int, out interface{}) error {
	if limit <= 0 {
		return s.connection.Find(out).Error
	}
	return s.connection.Limit(limit).Find(out).Error
}

func (s *sqlite) GetWhere(limit int, whereQuery, whereArgs, out interface{}) error {
	if limit <= 0 {
		return s.connection.Where(whereQuery, whereArgs).Find(out).Error
	}
	return s.connection.Limit(limit).Where(whereQuery, whereArgs).Find(out).Error
}

func (s *sqlite) GetWhereObject(where, out interface{}) error {
	return s.connection.Where(where).First(out).Error
}

func (s *sqlite) GetWhereOrderBy(limit int, orderBy string, whereQuery, whereArgs, out interface{}) error {
	if limit <= 0 {
		return s.connection.Where(whereQuery, whereArgs).Order(orderBy).Find(out).Error
	}
	return s.connection.Limit(limit).Where(whereQuery, whereArgs).Order(orderBy).Find(out).Error
}

func (s *sqlite) GetWhereOrderByWithMultipleArgs(limit int, orderBy string, out, whereQuery interface{}, whereArgs ...interface{}) error {
	if limit <= 0 {
		return s.connection.Where(whereQuery, whereArgs...).Order(orderBy).Find(out).Error
	}
	return s.connection.Limit(limit).Where(whereQuery, whereArgs...).Order(orderBy).Find(out).Error
}

// Update multiple attributes with `struct`, will only update those changed & non blank fields
func (s *sqlite) Set(out interface{}) error {
	return s.connection.Model(out).Updates(out).Error
}

// Similar to Set function but you can specify where clause to update rows
func (s *sqlite) SetWhere(whereQuery, whereArgs, out interface{}) error {
	return s.connection.Model(out).Where(whereQuery, whereArgs).Updates(out).Error
}

func (s *sqlite) Execute(out interface{}, sql string, values ...interface{}) error {
	tx := s.connection.Begin()

	if tx.Error != nil {
		return fmt.Errorf("failed to initialize transaction : %+v", tx.Error)
	}

	tx.Raw(sql, values...).Scan(out)
	return s.commit(tx)
}

func (s *sqlite) commit(tx *gorm.DB) error {
	err := tx.Error
	if err == nil {
		err = tx.Commit().Error
	}

	if err != nil {
		err = tx.Rollback().Error
		if err != nil {
			return fmt.Errorf("failed to rollback : %+v", err)
		}
		return fmt.Errorf("failed to perform operation : %+v", err)
	}
	return nil
}
