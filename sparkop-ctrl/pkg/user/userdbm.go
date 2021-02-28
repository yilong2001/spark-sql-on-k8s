package user

import (
  "fmt"
  //"conf"
  "time"
  "errors"
  //"gorm.io/driver/mysql"
  //"gorm.io/gorm"

  log "github.com/sirupsen/logrus"
  mmysql "github.com/spark-sql-on-k8s/pkg/mysql"
)

type UserDBManager struct {
	dbm *mmysql.DatabaseManager
}

var userDBM *UserDBManager = nil

func NewUserDBManager(dbm *mmysql.DatabaseManager) *UserDBManager {
	return &UserDBManager{dbm:dbm}
}

func InitUserDBM(dbm *mmysql.DatabaseManager) error {
	if (userDBM != nil) {
		return nil
	}

	userDBM = NewUserDBManager(dbm)
	userDBM.Init()

	dbm.RegisterTableModel(&User{})
	return nil
}

func (m *UserDBManager) Init() error {
	db := m.dbm.GetDB().Exec(CreateSparkUserInfoTable)
    if db == nil {
        msg := fmt.Sprintf("Failed to exec: '%s'", CreateSparkUserInfoTable)
        return errors.New(msg)
    }

    adminResult := db.Create(&User{Username: ADMIN_NAME, 
        Password: ADMIN_FIRST_PW,
        Uptime: time.Now()})
    log.Infof("create admin user result : %v", adminResult)
    return nil
}

// query
func (m *UserDBManager) GetUserInfo(username string) (*User, error) {
   var quser User
   m.dbm.GetDB().Table(getTableName(username)).
     Where("`username` = ?", username).
     First(&quser)

   if quser.Username == "" {
       return nil, fmt.Errorf("user(%s) not exists", username)
   }
   return &quser, nil
}

// create new user on db
func (m *UserDBManager) CreateUser(user User) (error, int64) {
    result := m.dbm.GetDB().Create(&user) // 通过数据的指针来创建

    //user.ID             // 返回插入数据的主键
    //result.Error        // 返回 error
    //result.RowsAffected // 返回插入记录的条数
    return result.Error, result.RowsAffected
}

func (m *UserDBManager) ValidateUserPassword(username, password string) (error, *User) {
    var user User
    result := m.dbm.GetDB().
      Where("`username` = ? AND `password` = ?", username, password).
      First(&user)
    //result := db.Table(getTableName(username)).Where(&User{Username:username, Password:password}).First(&user)

    log.Infof("validate db user password : %s, %s, %v", username, password, user)
    log.Info(result)

    if (user.ID > 0) {
        return nil, &user        
    }

    return fmt.Errorf("user(%s) or password is error", username), nil
}

func (m *UserDBManager) UpdatePassword(username string, password string) int64 {
   return m.dbm.GetDB().Model(&User{}).Where("`username` = ?", username).Updates(User{Password: password, Uptime: time.Now()}).RowsAffected
}


func (m *UserDBManager) UpdateMaxCores(username string, maxcores int32) int64 {
   return m.dbm.GetDB().Model(&User{}).Where("`username` = ?", username).Updates(User{Maxcores: maxcores, Uptime: time.Now()}).RowsAffected
}

// update max memory
func (m *UserDBManager) UpdateMaxMemory(username string, maxmem int32) int64 {
   return m.dbm.GetDB().Model(&User{}).Where("`username` = ?", username).Updates(User{Maxmemory: maxmem, Uptime: time.Now()}).RowsAffected
}

// update max cores and memory
func (m *UserDBManager) UpdateMaxCoresAndMemory(username string, maxcores int32, maxmem int32) int64 {
   return m.dbm.GetDB().Model(&User{}).Where("`username` = ?", username).Updates(User{Maxcores: maxcores, Maxmemory: maxmem, Uptime: time.Now()}).RowsAffected
}
