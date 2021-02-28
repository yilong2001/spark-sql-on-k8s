package controller

import (
	//"context"
	//"fmt"
	"time"
	"strconv"
	//"reflect"
	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"

	//sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	//myutil "github.com/spark-sql-on-k8s/sparkop-ctrl/sparkappctrl/util"

	muser "github.com/spark-sql-on-k8s/pkg/user"
)

type CreateUserReq struct {
	Username string   `form:"username" json:"username" xml:"username" binding:"required"`
	Password string   `form:"password" json:"password" xml:"password" binding:"required"`
	Sets     map[string]string `form:"sets" json:"sets" xml:"sets"`
}

func (r *SparkAppCtrl) RegisterUser(c *gin.Context) {
	var user muser.User
	var userReq CreateUserReq
	err := c.ShouldBind(&userReq)
    if err != nil {
    	log.Errorf("error : bind create user req error : %v!", err)
        buildErrorResponse(c, err)
        return
    }

    var maxCores int32 = 0
    var maxMem   int32 = 0
    if userReq.Sets != nil {
    	if maxCoresStr, ok := userReq.Sets["maxcores"]; ok {
    		if s, err := strconv.ParseInt(maxCoresStr, 10, 32); err == nil {
				maxCores = int32(s)
			}
    	}

    	if maxMemStr, ok := userReq.Sets["maxmemory"]; ok {
    		if s, err := strconv.ParseInt(maxMemStr, 10, 32); err == nil {
				maxMem = int32(s)
			}
    	}
    }

    err = muser.CreateUserInfo(muser.User{
    	Username: userReq.Username,
    	Password: userReq.Password,
    	Maxmemory: maxMem,
    	Maxcores: maxCores,
    	Uptime: time.Now(),
        })
    if err != nil {
    	log.Errorf("error : create user error : %v, %v!", user, err)
        buildErrorResponse(c, err)
        return
    }

    buildSuccessResponse(c, "create user ok!")
}

//func (r *SparkAppCtrl) UpdateMaxCores(c *gin.Context) {
//
//}

