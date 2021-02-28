package user

const (
    timeFormat = "2006-01-02 15:04:05"
    ADMIN_NAME = "admin"
    ADMIN_FIRST_PW = "admin"
)

const (
    CreateSparkUserInfoTable = `
    CREATE TABLE IF NOT EXISTS spark_userinfo (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) UNIQUE not null,
    password VARCHAR(64),
    maxcores INT,
    maxmemory INT,
    uptime DATETIME)
    `
)
