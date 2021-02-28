package user

import (
	"time"
)

// User gorm user object
type User struct {
    ID       int32       `gorm:"column:id;type:int(11);primary key"`
    Username string      `gorm:"column:username;type:varchar(64);unique;not null"`
    Password   string    `gorm:"column:password;type:varchar(64);not null"`

    Maxcores     int32   `gorm:"column:maxcores;type:int(11)"`
    // memory unit: M
    Maxmemory  int32     `gorm:"column:maxmemory;type:int(11)"`

    Uptime   time.Time   `gorm:"column:uptime;type:datetime"`
}

// TableName gorm use this to get tablename
func (u User) TableName() string {
    return "spark_userinfo"
}

// for table
func getTableName(username string) string {
    return "spark_userinfo"
}
