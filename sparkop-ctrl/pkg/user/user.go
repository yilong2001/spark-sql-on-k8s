package user

import (
    "fmt"
	log "github.com/sirupsen/logrus"
)

const (
    UpdatePassword = 1
    UpdateMaxCores = 2
    UpdateMaxMem   = 3
    UpdateMaxCoreAndMem = 4
)

func GetUserInfo(username string) (*User, error) {
    // try cache
    user, err := getUserCacheInfo(username)
    if err == nil && user.Username == username {
        return user, err
    }

    // get from db
    user, err = userDBM.GetUserInfo(username)
    if err != nil {
        return user, err
    }

    // update cache
    serr := setUserCacheInfo(user)
    if serr != nil {
        log.Error("cache userinfo failed for user:", user.Username, " with err:", serr.Error())
    }

    return user, err
}

func CreateUserInfo(user User) error {
    err, rows := userDBM.CreateUser(user)
    if err != nil {
        return err
    }

    if rows == 0 {
        return fmt.Errorf("create user failed, please try again after a minute")
    }

    updateCachedUserinfo(user)

    return nil
}

func ValidateUserPassword(username, password string) (error, *User) {
    return userDBM.ValidateUserPassword(username, password)
}

// edit userinfo
func UpdateUserInfo(username, password string, maxcores, maxmem int32, mode uint32) int64 {
    // update db info
    var affectedRows int64
    switch mode {
        case UpdatePassword:
            affectedRows = userDBM.UpdatePassword(username, password)
        case UpdateMaxCores:
            affectedRows = userDBM.UpdateMaxCores(username, maxcores)
        case UpdateMaxMem:
            affectedRows = userDBM.UpdateMaxMemory(username, maxmem)
        case UpdateMaxCoreAndMem:
            affectedRows = userDBM.UpdateMaxCoresAndMemory(username, maxcores, maxmem)
        default:
            // do nothing
            break
    }

    // on successing, update cache or delete it if updating failed
    if affectedRows == 1 {
        user, err := userDBM.GetUserInfo(username)
        if err == nil {
            updateCachedUserinfo(*user)
        } else {
            log.Errorf("Failed to get dbUserInfo for cache, username: %s  with err: %v", username, err)
        }
    }

    return affectedRows
}


