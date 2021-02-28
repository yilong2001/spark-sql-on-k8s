package main

import (
	"context"
	//"bytes"
	//"errors"
	//"plugin"
	//"flag"
	"fmt"
	"io/ioutil"
	//"os"
	//"os/signal"
	//"strings"
	//"syscall"
	//"time"
	//"path/filepath"
	//"github.com/satori/go.uuid"

	//"github.com/golang/glog"
	//"k8s.io/klog"
	"github.com/spf13/pflag"
	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/gin-gonic/gin"
    //"github.com/kataras/iris/v12"
	//"github.com/micro/go-micro/web"
	//"github.com/micro/micro/v3/web"
	//apiv1 "k8s.io/api/core/v1"

	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	myctrl "github.com/spark-sql-on-k8s/sparkop-ctrl/sparkappctrl/controller"
	myutil "github.com/spark-sql-on-k8s/sparkop-ctrl/sparkappctrl/util"
	mymw "github.com/spark-sql-on-k8s/sparkop-ctrl/plugin/middleware"
	myauth "github.com/spark-sql-on-k8s/sparkop-ctrl/plugin/auth"
	muser "github.com/spark-sql-on-k8s/pkg/user"
	mmysql "github.com/spark-sql-on-k8s/pkg/mysql"
)

func main() {
	sacommon.InitLogDefault()

	sparkCtrlConf := myutil.NewSparkCtrlConfig()
	sparkCtrlConf.AddFlags(pflag.CommandLine)
	pflag.Parse()

	sparkCtrlConf.SetLog()

	//sparkCtrlConf.MasterHost = ""
	//sparkCtrlConf.KubeConfig = "/var/lib/rancher/k3s/server/cred/admin.kubeconfig"
	//sparkCtrlConf.SparkAppNamespace = "spark-jobs"
	//sparkCtrlConf.LabelSelectorFilter = ""
	//sparkCtrlConf.ResyncInterval = 0
	//sparkCtrlConf.TmpLocalDir = "../logs"

	ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)

    dbm, err := mmysql.CreateDatabaseManager(&sparkCtrlConf.MetaDb)
    if err != nil {
    	panic(err)
    }
    defer dbm.CloseDB()

    muser.InitUserDBM(dbm)

	sparkCtrl := myctrl.NewSparkAppCtrl(sparkCtrlConf)
	defer sparkCtrl.Stop(ctx)

	stopCh := make(chan struct{}, 1)
	defer func() {
		cancel()
		close(stopCh)
	}()

	if err := sparkCtrl.Start(stopCh); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20  // 8 MiB

	router.Handle(http.MethodGet, "/hello", func(c *gin.Context) {
		sacommon.NewGinSuccessMessageResponse(nil,nil,"hello","",c)
	})

	router.POST("/api/v1/auth/jwt", func(ctx *gin.Context) {
		token, err := myauth.BuildLoginController(myauth.BuildLoginService(nil),
			myauth.BuildJWTService()).Login(ctx)
		if token != "" {
			sacommon.NewGinSuccessMessageResponse(gin.H{
				"token": token,
			},nil,"ok","",ctx)
		} else {
			sacommon.NewGinFaliResponse(http.StatusUnauthorized, fmt.Sprintf("%v",err),"",ctx)
		}
	})
	
	v1Group := router.Group("/api/v1")
	k8sRouter := v1Group.Group("/k8s")
	{
		k8sRouter.Use(mymw.AuthorizeJWT())
		k8sRouter.POST("/app/spark-app-yaml", 
			sparkCtrl.ApplyWithUploadedYaml)

		k8sRouter.POST("/app/spark-app", 
			sparkCtrl.ApplyWithUploadedJars)

		k8sRouter.POST("/app/spark-sql/engine", 
			sparkCtrl.StartSparkSqlEngineApplication)

		k8sRouter.GET("/app/spark-sql/engine", 
			sparkCtrl.GetSparkSqlEngineApplicationState)

		k8sRouter.DELETE("/app/spark-sql/engine", 
			sparkCtrl.ShutdownSparkSqlEngineApplication)

		k8sRouter.GET("/app/spark-app/ns", 
			sparkCtrl.ListApplication)

		k8sRouter.GET("/app/spark-app/ns/:ns", 
			sparkCtrl.ListApplication)

		k8sRouter.GET("/app/spark-app/info/ns", 
			sparkCtrl.ListApplicationInfo)

		k8sRouter.GET("/app/spark-app/info/ns/:ns", 
			sparkCtrl.ListApplicationInfo)

		k8sRouter.DELETE("/app/spark-app/ns/:ns/appname/:appname", 
			sparkCtrl.DeleteApplication)

		k8sRouter.GET("/app/ns/:ns/service/:service", 
			sparkCtrl.ServiceInfo)

		k8sRouter.GET("/app/ns/:ns/pod/:pod", 
			sparkCtrl.PodInfo)
	}

	userRoute := v1Group.Group("/user")
	{
		userRoute.Use(mymw.AuthorizeJWT())
		userRoute.POST("/register", sparkCtrl.RegisterUser)
	}

	router.POST("/api/v1/s3/upload/ns/:ns/path/:path", 
		sparkCtrl.UploadFilesToS3Path)

	router.POST("/upload", func(c *gin.Context) {
		var content_length int64
		content_length = c.Request.ContentLength
		if content_length <= 0 || content_length > int64(sparkCtrlConf.MaxUploadFileSize) {
			log.Infof("upload file length error : %v\n", content_length)
			c.String(http.StatusInternalServerError, fmt.Sprintf("upload file length error : %v", content_length))
			return
		}

		// single file
		file, _ := c.FormFile("file")
		log.Info(file.Filename)

		// Upload the file to specific dst.
		c.SaveUploadedFile(file, "./tmpfile")

		src, err := file.Open()
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
			return
		}
		defer src.Close()

		fileBytes, err := ioutil.ReadAll(src)
    	if err != nil {
        	fmt.Println(err)
    	}
    	
    	//log.Info(string(fileBytes))
		c.String(http.StatusOK, fmt.Sprintf("'%s : \n %s' uploaded!", file.Filename, string(fileBytes)))
	})

	router.Run(fmt.Sprintf("%s:%v", 
		sparkCtrlConf.WebBindHost, 
		sparkCtrlConf.WebBindPort))
}

