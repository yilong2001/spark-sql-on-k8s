package controller

import (
	//"context"
	//"bytes"
	//"errors"
	//"plugin"
	//"flag"
	"fmt"
	//"encoding/json" 
	"os"
	//"path/filepath"
	"strings"
	//"net/url"
	//"path/filepath"
	//"os/signal"
	//"reflect"
	//"syscall"
	//"time"

	//"github.com/golang/glog"
	//"k8s.io/klog"
	//"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	//"github.com/satori/go.uuid"
	"github.com/gin-gonic/gin"
	//"net/http"

	//"path/filepath"
	//"runtime"
	//"strconv"
	//"strings"

	//apiv1 "k8s.io/api/core/v1"
	//apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/util/clock"
	//"k8s.io/client-go/informers"
	//clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	//"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/client-go/tools/record"

	//crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	//crdinformers "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/informers/externalversions"
	//apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	//"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/config"

	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	//so "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
)

// TODO: 怎么支持 UDF ?
/*****************************************************
* 
*****************************************************/
const (
	DEFAULT_SQL_ENGINE_NAMESPACE = "spark-jobs"
	SPARK_SQL_ENGINE_JAR = "kyuubi-spark-sql-engine-1.0.0-SNAPSHOT.jar"
	SPARK_SQL_METADB_TYPE = "MYSQL" // or PG
	SPARK_SQL_DRIVER_JAR_MYSQL = ""
	SPARK_SQL_DRIVER_JAR_PG = ""

	DEFAULT_SPARK_SQL_IMAGE_NAME = "registry.cn-beijing.aliyuncs.com/yilong2001/spark:v3.0.1-1216"
)

/*****************************************************
* 
*****************************************************/
type SparkSqlParam struct {
	DriverCores int32 `form:"drivercores" json:"drivercores" xml:"drivercores"`
	DriverCoreLimit string `form:"drivercorelimit" json:"drivercorelimit" xml:"drivercorelimit"`
	DriverMem   string `form:"drivermem" json:"drivermem" xml:"drivermem"`

	ExectorCores int32 `form:"exectorcores" json:"exectorcores" xml:"exectorcores"`
	ExectorCoreLimit string `form:"exectorcorelimit" json:"exectorcorelimit" xml:"exectorcorelimit"`
	ExectorMem  string `form:"exectormem" json:"exectormem" xml:"exectormem"`

	ExtraConf    map[string]string `form:"extraconf" json:"extraconf" xml:"extraconf"`
}

func BuildDefaultSparkSqlParam() SparkSqlParam {
	return SparkSqlParam {
		DriverCores: 1,
		DriverCoreLimit: "1200m",
		DriverMem: "512m",
		ExectorCores: 1,
		ExectorCoreLimit: "1200m",
		ExectorMem: "512m",
	}
}

/*****************************************************
* 
*****************************************************/
func (r *SparkAppCtrl) buildSparkAppParamWithSqlParam(sqlParam SparkSqlParam) SparkAppParam {
	appParam := SparkAppParam {
		TimeToLiveSeconds: 60,

		SparkVersion: "3.0.1",
		ProxyUser: "root",

		DriverCores: sqlParam.DriverCores,
		DriverCoreLimit: sqlParam.DriverCoreLimit,
		DriverMem: sqlParam.DriverMem,

		ExectorCores: sqlParam.ExectorCores,
		ExectorCoreLimit: sqlParam.ExectorCoreLimit,
		ExectorMem: sqlParam.ExectorMem,

		ServiceAccount: "spark",
	}

	appParam.ExtraConf = make(map[string]string)

	appParam.ExtraConf["spark.thrift.jdbc.bind.port"] = "10019"
	appParam.ExtraConf["spark.hadoop.javax.jdo.option.ConnectionUserName"] = r.cfg.MetaDb.User
	appParam.ExtraConf["spark.hadoop.javax.jdo.option.ConnectionPassword"] = r.cfg.MetaDb.Password

	if r.cfg.MetaDb.Type == "MYSQL" {
		appParam.ExtraConf["spark.hadoop.javax.jdo.option.ConnectionDriverName"] = "com.mysql.jdbc.Driver"
		appParam.ExtraConf["spark.hadoop.javax.jdo.option.ConnectionURL"] = "jdbc:mysql://"+r.cfg.MetaDb.Host+":"+r.cfg.MetaDb.Port+"/"+r.cfg.MetaDb.Dbname+"?createDatabaseIfNotExist=true"
	} else {
		appParam.ExtraConf["spark.hadoop.javax.jdo.option.ConnectionDriverName"] = "org.postgresql.Driver"
		appParam.ExtraConf["spark.hadoop.javax.jdo.option.ConnectionURL"] = "jdbc:postgresql://"+r.cfg.MetaDb.Host+":"+r.cfg.MetaDb.Port+"/"+r.cfg.MetaDb.Dbname+"?createDatabaseIfNotExist=true"
	}

	// copy user defined spark conf to sql application
	if ((sqlParam.ExtraConf != nil) && len(sqlParam.ExtraConf) > 0) {
		for k,v := range sqlParam.ExtraConf {
			if strings.HasPrefix(k, "spark.") {
				appParam.ExtraConf[k] = v
			}
		}
	}


	return appParam
}


/*****************************************************
* 
*****************************************************/
func (r *SparkAppCtrl) StartSparkSqlEngineApplication(c *gin.Context) {
	userif, ok := c.Get("X-User")
	if !ok {
		log.Errorf("X-User does not exist")
		buildErrorResponse(c, fmt.Errorf("X-User does not exist"))
		return
	}
	user := userif.(string)
	if (user == "") {
		user = "default"
	}

	oldApp, err := r.crdClient.
		SparkoperatorV1beta2().
		SparkApplications(DEFAULT_SQL_ENGINE_NAMESPACE).
		Get("spark-sql-"+user, metav1.GetOptions{})
	if (err == nil && oldApp != nil) {
		oldSvc, err := r.kubeClient.CoreV1().
		Services(DEFAULT_SQL_ENGINE_NAMESPACE).
		Get(sacommon.GetSparkAppJDBCServiceName("spark-sql-"+user), metav1.GetOptions{})
		if err != nil {
			log.Errorf("error: spark-sql-%s 's service wrong : %v!", user, err)
			buildErrorResponse(c, err)
			return
		}

		sacommon.NewGinSuccessMessageResponse(r.getSparkService(oldSvc), nil, "success","", c)
		return
	}

	sqlParam := BuildDefaultSparkSqlParam()
	//var sqlParam SparkSqlParam

	err = c.ShouldBind(&sqlParam)
    if err != nil {
    	log.Errorf("error : bind SparkSqlParam %v!", err)
        buildErrorResponse(c, err)
        return
    }
    log.Infof("*** *** sql params : %v", sqlParam)

    sqlParam.DriverCoreLimit = fmt.Sprintf("%dm", sqlParam.DriverCores * 1100)
    sqlParam.ExectorCoreLimit = fmt.Sprintf("%dm", sqlParam.ExectorCores * 1100)

	sqlEngineJar := os.Getenv("SPARK_SQL_ENGINE_JAR")
	if (sqlEngineJar == "") {
		log.Errorf("error: start default spark sql app : SPARK_SQL_ENGINE_JAR not set")
		buildErrorResponse(c, fmt.Errorf("error: start default spark sql app : SPARK_SQL_ENGINE_JAR not set"))
		return
	}
	
	dbType := strings.ToUpper(r.cfg.MetaDb.Type)
	if (dbType == "") {
		dbType = "MYSQL"
		log.Errorf("error: start default spark sql app : SPARK_SQL_METADB_TYPE not set, use MYSQL")
	}

	sqlDriverJar := os.Getenv("SPARK_SQL_DRIVER_JAR_" + dbType)
	if (sqlDriverJar == "") {
		log.Errorf("error: start default spark sql app : %s not set", "SPARK_SQL_DRIVER_JAR_" + dbType)
		buildErrorResponse(c, fmt.Errorf("error: start default spark sql app : %s not set", "SPARK_SQL_DRIVER_JAR_" + dbType))
		return
	}

	sparkImageName := os.Getenv("SPARK_SQL_IMAGE_NAME")
	if (sparkImageName == "") {
		sparkImageName = DEFAULT_SPARK_SQL_IMAGE_NAME
	}

	// TODO: 暂不考虑动态更新 jar 文件的情况
	if (!r.isSqlJarUploaded) {
		s3files, err := r.doUploadToS3(DEFAULT_SQL_ENGINE_NAMESPACE, "jars", []string{sqlEngineJar, sqlDriverJar})
		if err != nil {
			log.Errorf("error: start default spark sql app , upload files error :%v", err)
			buildErrorResponse(c, err)
			return
		}

		r.sqlEngineS3Path = s3files[0]
		r.sqlDriverS3Path = s3files[1]
	}

	appParam := r.buildSparkAppParamWithSqlParam(sqlParam)
	appParam.MainFile = r.sqlEngineS3Path
	appParam.MainClass = "org.apache.kyuubi.engine.spark.SparkSQLEngine"
	appParam.Image = sparkImageName
	appParam.ImagePolicy = "IfNotPresent"

	appParam.JarFiles = []string{r.sqlDriverS3Path}
	appParam.Namespace = DEFAULT_SQL_ENGINE_NAMESPACE
	appParam.Name = "spark-sql-" + user

	log.Infof("Application Param from Spark Sql Request : %v", appParam)

	r.startApplicationFromParam(c, appParam)
}

func (r *SparkAppCtrl) ShutdownSparkSqlEngineApplication(c *gin.Context) {
	userif, ok := c.Get("X-User")
	if !ok {
		log.Errorf("X-User does not exist")
		buildErrorResponse(c, fmt.Errorf("X-User does not exist"))
		return
	}

	user := userif.(string)
	if (user == "") {
		user = "default"
	}
	appname := "spark-sql-" + user

	err := deleteSparkApplication(DEFAULT_SQL_ENGINE_NAMESPACE, appname, r.crdClient)
	if err != nil {
		log.Errorf("error : delete spark app %s/%s error : %v !", DEFAULT_SQL_ENGINE_NAMESPACE, appname, err)
		buildErrorResponse(c, err)
		return
	}

	buildSuccessResponse(c, fmt.Sprintf("'%s/%s' resource deleted!", DEFAULT_SQL_ENGINE_NAMESPACE, appname))
}

func (r *SparkAppCtrl) GetSparkSqlEngineApplicationState(c *gin.Context) {
	userif, ok := c.Get("X-User")
	if !ok {
		log.Errorf("X-User does not exist")
		buildErrorResponse(c, fmt.Errorf("X-User does not exist"))
		return
	}

	user := userif.(string)
	if (user == "") {
		user = "default"
	}
	appname := "spark-sql-" + user

	app, err := r.crdClient.SparkoperatorV1beta2().
		SparkApplications(DEFAULT_SQL_ENGINE_NAMESPACE).
		Get(appname, metav1.GetOptions{})
	if err != nil {
		log.Errorf("get spark sql application error, %v", err)
		buildErrorResponse(c, fmt.Errorf("get spark sql application error, %v", err))
		return
	}

	appState := sacommon.SparkAppState {
		Namespace: app.Namespace,
		Name: app.Name,
		State: sacommon.StateMessage {
			State: string(app.Status.AppState.State),
			Message: app.Status.AppState.ErrorMessage,
		},
	}
	
	sacommon.NewGinSuccessMessageResponse(appState, nil, "success","", c)
}
