package controller

import (
	//"context"
	//"bytes"
	//"errors"
	//"plugin"
	//"flag"
	"fmt"
	"encoding/json" 
	//"os"
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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	//"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/config"

	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	//so "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
)

var hostPathDirectoryOrCreate apiv1.HostPathType = apiv1.HostPathDirectoryOrCreate
var hostPathDirectory apiv1.HostPathType = apiv1.HostPathDirectory

type SparkAppParam struct {
	TimeToLiveSeconds int64 `form:"ttls" json:"ttls" xml:"ttls"`
	Name string   `form:"name" json:"name" xml:"name" binding:"required"`
	Namespace string `form:"namespace" json:"namespace" xml:"namespace" binding:"required"`
	MainClass string `form:"mainclass" json:"mainclass" xml:"mainclass" binding:"required"`
	MainFile  string `form:"mainfile" json:"mainfile" xml:"mainfile" binding:"required"`
	Arguments []string `form:"arguments" json:"arguments" xml:"arguments"`

	Image string `form:"image" json:"image" xml:"image" binding:"required"`
	ImagePolicy string `form:"imagepolicy" json:"imagepolicy" xml:"imagepolicy" binding:"required"`

	JarFiles  []string `form:"jarfiles" json:"jarfiles" xml:"jarfiles"`

	ProxyUser string `form:"proxyuser" json:"proxyuser" xml:"proxyuser"`
	ServiceAccount string `form:"serviceaccount" json:"serviceaccount" xml:"serviceaccount"`

	SparkVersion string `form:"sparkversion" json:"sparkversion" xml:"sparkversion"`
	
	DriverCores int32 `form:"drivercores" json:"drivercores" xml:"drivercores"`
	DriverCoreLimit string `form:"drivercorelimit" json:"drivercorelimit" xml:"drivercorelimit"`
	DriverMem   string `form:"drivermem" json:"drivermem" xml:"drivermem"`

	ExectorCores int32 `form:"exectorcores" json:"exectorcores" xml:"exectorcores"`
	ExectorCoreLimit string `form:"exectorcorelimit" json:"exectorcorelimit" xml:"exectorcorelimit"`
	ExectorMem  string `form:"exectormem" json:"exectormem" xml:"exectormem"`

	EventLogEnable string `form:"eventlogenable" json:"eventlogenable" xml:"eventlogenable"`
	EventLogDir    string `form:"eventlogdir" json:"eventlogdir" xml:"eventlogdir"`
	
	S3Endpoint  string `form:"s3endpoint" json:"s3endpoint" xml:"s3endpoint"`
	S3AccessKey string `form:"s3accesskey" json:"s3accesskey" xml:"s3accesskey"`
	S3SecretKey string `form:"s3secretkey" json:"s3secretkey" xml:"s3secretkey"`

	//ExtraConf   []KVPair `form:"extraconf" json:"extraconf" xml:"extraconf"`
	ExtraConf     map[string]string `form:"extraconf" json:"extraconf" xml:"extraconf"`
}

func BuildDefaultSparkAppParam() SparkAppParam {
	return SparkAppParam {
		TimeToLiveSeconds: 60,

		SparkVersion: "3.0.1",

		DriverCores: 1,
		DriverCoreLimit: "1200m",
		DriverMem: "512m",

		ExectorCores: 1,
		ExectorCoreLimit: "1200m",
		ExectorMem: "512m",

		ServiceAccount: "spark",
	}
}


func (r *SparkAppCtrl) buildSparkConfig(appParam SparkAppParam) map[string]string {
	sparkConf := make(map[string]string)
	sparkConf["spark.sql.parquet.mergeSchema"] = "false"
	sparkConf["spark.sql.parquet.filterPushdown"] = "true"
	sparkConf["spark.sql.hive.metastorePartitionPruning"] = "true"

	sparkConf["spark.hadoop.fs.s3a.impl"] = "org.apache.hadoop.fs.s3a.S3AFileSystem"
	sparkConf["spark.hadoop.fs.s3a.impl.disable.cache"] = "true"
	sparkConf["spark.hadoop.fs.s3a.bucket.probe"] = "0"
	sparkConf["spark.hadoop.fs.s3a.connection.ssl.enabled"] = "false"
	sparkConf["spark.hadoop.fs.s3a.committer.staging.conflict-mode"] = "append"
	sparkConf["spark.hadoop.fs.hdfs.impl"] = "org.apache.hadoop.hdfs.DistributedFileSystem"
	sparkConf["spark.hadoop.fs.file.impl"] = "org.apache.hadoop.fs.LocalFileSystem"
	
	if appParam.S3Endpoint != "" {
		sparkConf["spark.hadoop.fs.s3a.endpoint"] = appParam.S3Endpoint
		sparkConf["spark.hadoop.fs.s3a.access.key"] = appParam.S3AccessKey
		sparkConf["spark.hadoop.fs.s3a.secret.key"] = appParam.S3SecretKey
	} else {
		sparkConf["spark.hadoop.fs.s3a.endpoint"] = r.cfg.S3Endpoint
		sparkConf["spark.hadoop.fs.s3a.access.key"] = r.cfg.S3AccessKey
		sparkConf["spark.hadoop.fs.s3a.secret.key"] = r.cfg.S3SecretKey
	}

	sparkConf["spark.hadoop.fs.s3a.aws.credentials.provider"] = "org.apache.hadoop.fs.s3a.TemporaryAWSCredentialsProvider,org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider,org.apache.hadoop.fs.s3a.auth.IAMInstanceCredentialsProvider"
	
	sparkConf["spark.hadoop.hive.metastore.schema.verification"] = "false"
	sparkConf["spark.hadoop.datanucleus.schema.autoCreateAll"] = "true"
	sparkConf["spark.hadoop.metastore.task.threads.always"] = "org.apache.hadoop.hive.metastore.events.EventCleanerTask"
	sparkConf["spark.hadoop.metastore.expression.proxy"] = "org.apache.hadoop.hive.metastore.DefaultPartitionExpressionProxy"

	if ((appParam.ExtraConf != nil) && len(appParam.ExtraConf) > 0) {
		for k,v := range appParam.ExtraConf {
			if strings.HasPrefix(k, "spark.") {
				sparkConf[k] = v
			}
		}
	}

	return sparkConf
}

func (r *SparkAppCtrl) buildSparkApplication(appParam SparkAppParam) *v1beta2.SparkApplication {
	sparkConf := make(map[string]string)
	sparkConf["spark.sql.parquet.mergeSchema"] = "false"
	sparkConf["spark.sql.parquet.filterPushdown"] = "true"
	sparkConf["spark.sql.hive.metastorePartitionPruning"] = "true"

	sparkConf["spark.hadoop.fs.s3a.impl"] = "org.apache.hadoop.fs.s3a.S3AFileSystem"
	sparkConf["spark.hadoop.fs.s3a.impl.disable.cache"] = "true"
	sparkConf["spark.hadoop.fs.s3a.bucket.probe"] = "0"
	sparkConf["spark.hadoop.fs.s3a.connection.ssl.enabled"] = "false"
	sparkConf["spark.hadoop.fs.s3a.committer.staging.conflict-mode"] = "append"
	sparkConf["spark.hadoop.fs.hdfs.impl"] = "org.apache.hadoop.hdfs.DistributedFileSystem"
	sparkConf["spark.hadoop.fs.file.impl"] = "org.apache.hadoop.fs.LocalFileSystem"
	
	if appParam.S3Endpoint != "" {
		sparkConf["spark.hadoop.fs.s3a.endpoint"] = appParam.S3Endpoint
		sparkConf["spark.hadoop.fs.s3a.access.key"] = appParam.S3AccessKey
		sparkConf["spark.hadoop.fs.s3a.secret.key"] = appParam.S3SecretKey
	} else {
		sparkConf["spark.hadoop.fs.s3a.endpoint"] = r.cfg.S3Endpoint
		sparkConf["spark.hadoop.fs.s3a.access.key"] = r.cfg.S3AccessKey
		sparkConf["spark.hadoop.fs.s3a.secret.key"] = r.cfg.S3SecretKey
	}

	sparkConf["spark.hadoop.fs.s3a.aws.credentials.provider"] = "org.apache.hadoop.fs.s3a.TemporaryAWSCredentialsProvider,org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider,org.apache.hadoop.fs.s3a.auth.IAMInstanceCredentialsProvider"
	
	sparkConf["spark.hadoop.hive.metastore.schema.verification"] = "false"
	sparkConf["spark.hadoop.datanucleus.schema.autoCreateAll"] = "true"
	sparkConf["spark.hadoop.metastore.task.threads.always"] = "org.apache.hadoop.hive.metastore.events.EventCleanerTask"
	sparkConf["spark.hadoop.metastore.expression.proxy"] = "org.apache.hadoop.hive.metastore.DefaultPartitionExpressionProxy"

	if ((appParam.ExtraConf != nil) && len(appParam.ExtraConf) > 0) {
		for k,v := range appParam.ExtraConf {
			if strings.HasPrefix(k, "spark.") {
				sparkConf[k] = v
			}
		}
	}

	labels := make(map[string]string)
	labels["version"] = appParam.SparkVersion

	app := &v1beta2.SparkApplication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: appParam.Namespace,
			Name:      appParam.Name,
			/*OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1beta2.SchemeGroupVersion.String(),
					Kind:       reflect.TypeOf(v1beta2.SparkApplication{}).Name(),
					Name:       sapp.Name,
					UID:        sapp.UID,
				},
			},*/
		},
		Spec: v1beta2.SparkApplicationSpec {
			TimeToLiveSeconds: &appParam.TimeToLiveSeconds,
			Type: v1beta2.ScalaApplicationType,
			SparkVersion: appParam.SparkVersion,
			Mode: v1beta2.ClusterMode,
			ProxyUser: &appParam.ProxyUser,
			Image: &appParam.Image,
			ImagePullPolicy: &appParam.ImagePolicy,
			MainClass: &appParam.MainClass,
			MainApplicationFile: &appParam.MainFile,
			Arguments: appParam.Arguments,
			SparkConf: sparkConf,

			Volumes: []apiv1.Volume{
				apiv1.Volume{
					Name: "tmp-volume",
					VolumeSource: apiv1.VolumeSource{
						HostPath: &apiv1.HostPathVolumeSource{
							Path: "/tmp",
							Type: &hostPathDirectory,
						},
					},
				},
			},

			Driver: v1beta2.DriverSpec{
				SparkPodSpec: v1beta2.SparkPodSpec{
					Cores: &appParam.DriverCores,
					CoreLimit: &appParam.DriverCoreLimit,
					Memory: &appParam.DriverMem,
					Labels: labels,
					ServiceAccount: &appParam.ServiceAccount,
					VolumeMounts: []apiv1.VolumeMount{
						apiv1.VolumeMount{
							Name: "tmp-volume",
							MountPath: "/tmp",
						},
					},
				},
			},
			Executor: v1beta2.ExecutorSpec{
				SparkPodSpec: v1beta2.SparkPodSpec{
					Cores: &appParam.ExectorCores,
					CoreLimit: &appParam.ExectorCoreLimit,
					Memory: &appParam.ExectorMem,
					Labels: labels,
					VolumeMounts: []apiv1.VolumeMount{
						apiv1.VolumeMount{
							Name: "tmp-volume",
							MountPath: "/tmp",
						},
					},
				},
			},
			Deps: v1beta2.Dependencies{
				Jars: appParam.JarFiles,
				ExcludePackages: []string {
					"org.apache.curator:curator-client",
					"com.google.guava:guava",
				},
			},
			RestartPolicy: v1beta2.RestartPolicy{
				Type: v1beta2.Never,
			},
		},
	}

	return app
}

func (r *SparkAppCtrl) ApplyWithUploadedJars(c *gin.Context) {
	log.Infof("gin context : %v", c)
	appParam := BuildDefaultSparkAppParam()

	err := c.ShouldBind(&appParam)
    if err != nil {
    	log.Errorf("error : bind AppParam %v!", err)
        buildErrorResponse(c, err)
        return
    }

	r.startApplicationFromParam(c, appParam)
}

func (r *SparkAppCtrl) startApplicationFromParam(c *gin.Context, appParam SparkAppParam) {
	app := r.buildSparkApplication(appParam)

	afterApp, err := createSparkApplication(appParam.Namespace, app, r.kubeClient, r.crdClient)

	if err != nil {
		log.Errorf("create spark by app : %v, \n error : %v", app, err)
		buildErrorResponse(c, fmt.Errorf("failed to create SparkApplication %s: %v", app.Name, err))
		return
	}

	svc,err := r.createSparkJDBCService(afterApp)
	if err != nil {
		log.Errorf("spark application is not spark-sql : %v", err)
	} else {
		log.Infof("spark sql service : %v", svc)
	}

	msg := sacommon.SparkAppCreatedMsg{
		Name: afterApp.Name,
		Namespace: afterApp.Namespace,
		UID: string(afterApp.UID),
		ExtraServices: make([]*sacommon.SparkService, 0),
	}

	if svc != nil {
		msg.ExtraServices = append(msg.ExtraServices, svc)
	}

	uisvc, _ := getDriverUIService(afterApp)
	if uisvc != nil {
		msg.ExtraServices = append(msg.ExtraServices, uisvc)
	}

	bytes, _ := json.Marshal(msg)
	r.natsStreamClient.SendMessage(sacommon.NatsChannelTraefik, string(bytes))

	sacommon.NewGinSuccessMessageResponse(svc, nil, "success","", c)	
}

/*****************************************************

*****************************************************/
