package mysql

import (
  "sync"

  "fmt"
  //"conf"
  //"time"
  "errors"
  "gorm.io/driver/mysql"
  "gorm.io/gorm"

  log "github.com/sirupsen/logrus"
  sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
)

//Manager db manager
type DatabaseManager struct {
	db      *gorm.DB
	config  *sacommon.DBConnConfig
	initOne sync.Once
	models  []TableModelInterface
}

const (
    Db_Conn_Maxidle = 50
    Db_Conn_Maxopen = 100
)


//CreateManager create manager
func CreateDatabaseManager(config  *sacommon.DBConnConfig) (*DatabaseManager, error) {
	log.Infof("crate database mysql config : %v", config)
	var db *gorm.DB
    var err error

	conninfo := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", 
        config.User, config.Password, config.Host, 
        config.Port, config.Dbname)
    
    db, err = gorm.Open(mysql.Open(conninfo), &gorm.Config{})
    if err != nil {
        msg := fmt.Sprintf("Failed to connect to db '%s', err: %s", conninfo, err.Error())
        return nil, errors.New(msg)
    }

	manager := &DatabaseManager{
		db:      db,
		config:  config,
		initOne: sync.Once{},
	}

	//db.SetLogger(manager)

	sqlDB, err := db.DB()
    if err != nil {
        msg := fmt.Sprintf("Failed to get db '%s', err: %s", conninfo, err.Error())
        return nil, errors.New(msg)
    }

    sqlDB.SetMaxIdleConns(Db_Conn_Maxidle)
    sqlDB.SetMaxOpenConns(Db_Conn_Maxopen)

	log.Info("mysql db driver create")
	return manager, nil
}

//CloseManager 关闭管理器
func (m *DatabaseManager) CloseDB() error {
    sqlDB, err := m.db.DB()
    if err != nil {
        log.Errorf("Close, but failed to get db err: %v", err)
        return nil
    }
    return sqlDB.Close()
}

//Begin begin a transaction
func (m *DatabaseManager) BeginTx() *gorm.DB {
	return m.db.Begin()
}

// EnsureEndTransactionFunc -
func (m *DatabaseManager) EnsureEndTransactionFunc() func(tx *gorm.DB) {
	return func(tx *gorm.DB) {
		if r := recover(); r != nil {
			log.Errorf("Unexpected panic occurred, rollback transaction: %v", r)
			tx.Rollback()
		}
	}
}

//Print Print
func (m *DatabaseManager) Print(v ...interface{}) {
	log.Info(v...)
}

//RegisterTableModel register table model
func (m *DatabaseManager) RegisterTableModel(model TableModelInterface) {
	m.models = append(m.models, model)
}

func (m *DatabaseManager) GetDB() *gorm.DB {
	return m.db
}

