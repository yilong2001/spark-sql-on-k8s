package controller

import (
	"context"
	//"bytes"
	//"errors"
	//"plugin"
	//"flag"
	"fmt"
	"encoding/json" 
	"os"
	"net/url"
	"path/filepath"
	//"os/signal"
	//"strings"
	//"syscall"
	//"time"

	//"github.com/golang/glog"
	//"k8s.io/klog"
	log "github.com/sirupsen/logrus"
	"github.com/satori/go.uuid"
	"github.com/gin-gonic/gin"
	"net/http"

	//"path/filepath"
	//"runtime"
	//"strconv"
	//"strings"

	//apiv1 "k8s.io/api/core/v1"
	//apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/util/clock"
	//"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	//"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/client-go/tools/record"

	crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
	//crdinformers "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/informers/externalversions"
	//apiv1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"

	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	//so "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	myutil "github.com/spark-sql-on-k8s/sparkop-ctrl/sparkappctrl/util"
)

const natsClientID  = "sparkapp-ctrl"

const defaultSparkUploadRoot = "spark-app-dependencies"

type  SparkAppCtrl struct {
	cfg           *myutil.SparkCtrlConfig
	kubeClient    clientset.Interface
	crdClient     crdclientset.Interface
	natsStreamClient *sacommon.NatsStreamingClient

	isSqlJarUploaded bool
	sqlEngineS3Path  string
	sqlDriverS3Path  string
}

func buildErrorResponse(c *gin.Context, err error) {
	sacommon.NewGinFaliResponse(http.StatusInternalServerError, 
		fmt.Sprintf("error: %v", err), "", c)
}

func buildSuccessResponse(c *gin.Context, msg string) {
	sacommon.NewGinSuccessMessageResponse(nil, nil, 
		msg, "", c)
}

/***********************************************
*
***********************************************/
func NewSparkAppCtrl(sparkCtrlCfg  *myutil.SparkCtrlConfig) *SparkAppCtrl {
	config, err := sacommon.BuildK8sRestConfig(sparkCtrlCfg.MasterHost, sparkCtrlCfg.KubeConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("build rest config ok! ")
	//log.Info("build rest config ok! ")

	kubeclient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("build kube clientset config ok! ")

	crdclient, err := crdclientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("build sparkapp crd clientset config ok! ")

	return &SparkAppCtrl{
		cfg: sparkCtrlCfg,
		kubeClient: kubeclient,
		crdClient: crdclient,
		natsStreamClient: sacommon.NewNatsStreamingClient(),
		isSqlJarUploaded: false,
	}
}

func (r *SparkAppCtrl) Start(stopCh <-chan struct{}) error {
	r.natsStreamClient.Start(stopCh, r.cfg.NatsStreamUrl, natsClientID)
	return nil
}

func (r *SparkAppCtrl) Stop(ctx context.Context) {
    r.natsStreamClient.Stop(ctx)
}

func (r *SparkAppCtrl) createSparkApplicationResource(yamlFile string) error {
	app, err := createFromYaml(r.cfg.SparkAppNamespace, yamlFile, r.kubeClient, r.crdClient)
	if err != nil {
		log.Errorf(" create spark application resource err : %v ", err)
		return err
	}

	svc,err := r.createSparkJDBCService(app)
	if err != nil {
		log.Errorf("spark application is not spark-sql : %v", err)
	} else {
		log.Infof("spark sql service : %v", svc)
	}

	msg := sacommon.SparkAppCreatedMsg{
		Name: app.Name,
		Namespace: app.Namespace,
		UID: string(app.UID),
		ExtraServices: make([]*sacommon.SparkService, 0),
	}

	if svc != nil {
		msg.ExtraServices = append(msg.ExtraServices, svc)
	}

	uisvc, _ := getDriverUIService(app)
	if uisvc != nil {
		msg.ExtraServices = append(msg.ExtraServices, uisvc)
	}

	bytes, _ := json.Marshal(msg)
	r.natsStreamClient.SendMessage(sacommon.NatsChannelTraefik, string(bytes))
	return nil
}

func (r *SparkAppCtrl) uploadLocalDependencies(appParam SparkAppParam, srcfile string) (string, error) {
	if r.cfg.S3UploadDir == "" {
		return "", fmt.Errorf("ERROR : unable to upload local dependencies: no s3 upload location specified!")
	}

	uploadLocationUrl, err := url.Parse(r.cfg.S3UploadDir)
	if err != nil {
		return "", err
	}
	uploadBucket := uploadLocationUrl.Host

	var uh *uploadHandler
	ctx := context.Background()
	switch uploadLocationUrl.Scheme {
	case "s3":
		uh, err = newS3Blob(ctx, uploadBucket, r.cfg.S3Endpoint, "")
	default:
		return "", fmt.Errorf("unsupported upload location URL scheme: %s", uploadLocationUrl.Scheme)
	}

	// Check if bucket has been successfully setup
	if err != nil {
		return "", err
	}

	uploadPath := filepath.Join(defaultSparkUploadRoot, appParam.Namespace, appParam.Name)

	uploadFilePath, err := uh.uploadToBucket(uploadPath, srcfile, false)

	return uploadFilePath, err
}

func (r *SparkAppCtrl) ApplyWithUploadedYaml(c *gin.Context) {
	log.Infof("gin context : %v", c)
	var content_length int64
	content_length = c.Request.ContentLength
	if content_length <= 0 || content_length > int64(r.cfg.MaxUploadFileSize) {
		log.Infof("upload file length error : %v\n", content_length)
		buildErrorResponse(c, fmt.Errorf("upload file length error : %v", content_length))
		return
	}

	yaml, err := c.FormFile("yaml")
	if err != nil {
    	log.Errorf("create spark app resource error : %v", err)
		buildErrorResponse(c, err)
		return			
	}

	ru := uuid.NewV4()
	tmpPath := fmt.Sprintf("%s/%s_%s", r.cfg.TmpLocalDir,
		yaml.Filename, ru)
	log.Infof("\nsrc file : %s --> local file : %s \n", yaml.Filename, tmpPath)
	c.SaveUploadedFile(yaml, tmpPath)

	err = r.createSparkApplicationResource(tmpPath)
	if err != nil {
    	log.Errorf("create spark app resource error : %v", err)
		buildErrorResponse(c, err)
		return
	}

	buildSuccessResponse(c, fmt.Sprintf("'%s' resource created!", yaml.Filename))
}

func (r *SparkAppCtrl) UploadFilesToS3Path(c *gin.Context) {
	s3ns := c.Param("ns")
	s3path := c.Param("path")

	if r.cfg.S3UploadDir == "" {
		err := fmt.Errorf("ERROR : unable to upload files to s3: no s3 upload location specified!")
		log.Errorf("%v", err)
    	buildErrorResponse(c, err)
		return
	}

	uploadLocationUrl, err := url.Parse(r.cfg.S3UploadDir)
	if err != nil {
		log.Errorf("%v", err)
    	buildErrorResponse(c, err)
		return
	}

	uploadBucket := uploadLocationUrl.Host
	log.Infof("upload file to s3 : %s, %s, %s \n", uploadLocationUrl, uploadLocationUrl.Scheme, uploadBucket)

	var uh *uploadHandler
	ctx := context.Background()
	switch uploadLocationUrl.Scheme {
	case "s3":
		uh, err = newPrivateS3Blob(ctx, uploadBucket, 
			r.cfg.S3Endpoint, r.cfg.S3AccessKey, r.cfg.S3SecretKey, "")
	default:
		err = fmt.Errorf("unsupported upload location URL scheme: %s", uploadLocationUrl.Scheme)
		log.Errorf("%v", err)
    	buildErrorResponse(c, err)
    	return
	}

	// Check if bucket has been successfully setup
	if err != nil {
		log.Errorf("%v", err)
    	buildErrorResponse(c, err)
    	return
	}

	var tmpFileArray []string = make([]string, 0)
	defer func() {
		for _, tmpFile := range tmpFileArray {
			os.Remove(tmpFile)
		}
	}()

	form, _ := c.MultipartForm()
    files := form.File["files[]"]
    if (len(files) == 0) {
    	log.Error("error : upload files is empty!")
    	buildErrorResponse(c, fmt.Errorf("error : upload files is empty !"))
		return
	}

	filesOnS3 := make(map[string]string)
	for _, file := range files {
		tmpPath := fmt.Sprintf("%s/%s", r.cfg.TmpLocalDir, file.Filename)
		log.Infof("save upload file : %s --> to local : %s ", file.Filename, tmpPath)
		c.SaveUploadedFile(file, tmpPath)
		tmpFileArray = append(tmpFileArray, tmpPath)

		uploadPath := filepath.Join(defaultSparkUploadRoot, s3ns, s3path)

		uploadFilePath, err := uh.uploadToBucket(uploadPath, tmpPath, true)
		if err != nil {
			log.Errorf("error : upload local file to s3 failed : %v !", err)
			buildErrorResponse(c, err)
			return
		}
		filesOnS3[file.Filename] = uploadFilePath
	}

	sacommon.NewGinSuccessMessageResponse(filesOnS3, nil, "success","", c)
}

/*****************************************************

*****************************************************/
func (r *SparkAppCtrl) DeleteApplication(c *gin.Context) {
	ns := c.Param("ns")
	appname := c.Param("appname")

	err := deleteSparkApplication(ns, appname, r.crdClient)
	if err != nil {
		log.Errorf("error : delete spark app %s/%s error : %v !", ns, appname, err)
		buildErrorResponse(c, err)
		return
	}

	buildSuccessResponse(c, fmt.Sprintf("'%s/%s' resource deleted!", ns, appname))
}

func (r *SparkAppCtrl) ListApplication(c *gin.Context) {
	ns := c.Param("ns")

	applist, err := listApplication(ns, r.crdClient)
	if err != nil {
		log.Errorf("error : list spark app %s error : %v !", ns, err)
		buildErrorResponse(c, err)
		return
	}

	if applist == nil || applist.Items == nil {
		log.Errorf("error : list spark app %s empty !", ns)
		buildErrorResponse(c, fmt.Errorf("error : list spark app %s empty !", ns))
		return
	}

	list := make([]interface{},0)
	for _, item := range applist.Items {
		list = append(list, item)
	}

	sacommon.NewGinSuccessMessageResponse(nil, list, "success","", c)
}

func (r *SparkAppCtrl) ListApplicationInfo(c *gin.Context) {
	ns := c.Param("ns")

	applist, err := listApplication(ns, r.crdClient)
	if err != nil {
		log.Errorf("error : list spark app %s error : %v !", ns, err)
		buildErrorResponse(c, err)
		return
	}

	if applist == nil || applist.Items == nil {
		log.Errorf("error : list spark app %s empty !", ns)
		buildErrorResponse(c, fmt.Errorf("error : list spark app %s empty !", ns))
		return
	}

	list := make([]interface{},0)
	for _, app := range applist.Items {
		uiurl, _ := sacommon.GetSparkAppRoutePath4UI(app.Namespace, app.Name)
		if r.cfg.UIEnableHttps {
			uiurl = "https://spark." + r.cfg.DomainName + uiurl
		} else {
			uiurl = "http://spark." + r.cfg.DomainName + uiurl
		}

		info := SparkAppInfo {
			Namespace: app.Namespace,
			Name: app.Name,
			State: string(app.Status.AppState.State),
			SubmissionAge: getSinceTime(app.Status.LastSubmissionAttemptTime),
			TerminationAge: getSinceTime(app.Status.TerminationTime),
			DriverUrl: uiurl,
		}

		_, ok := app.Spec.SparkConf[sparkJdbcPortConfigurationKey]
		if ok {
			jdbcsvcname := sacommon.GetSparkAppJDBCServiceName(app.Name)
			info.JdbcSvc = "jdbc:hive2://" + jdbcsvcname + ":10019"
		} else {
			info.JdbcSvc = "null"
		}

		list = append(list, info)
	}

	sacommon.NewGinSuccessMessageResponse(nil, list, "success","", c)
}

/*****************************************************
* 
*****************************************************/
func (r *SparkAppCtrl) ServiceInfo(c *gin.Context) {
	ns := c.Param("ns")
	service := c.Param("service")

	ctx := context.TODO()
	serviceInfo, err := r.getSparkServiceStatus(ctx, ns, service)
	if err != nil {
		log.Errorf("error : get spark service %s/%s status error : %v !", ns, service, err)
		buildErrorResponse(c, err)
		return
	}

	sacommon.NewGinSuccessMessageResponse(serviceInfo, nil, "success","", c)
}

func (r *SparkAppCtrl) PodInfo(c *gin.Context) {
	ns := c.Param("ns")
	pod := c.Param("pod")

	//ctx := context.TODO()
	podInfo, err := r.kubeClient.CoreV1().Pods(ns).Get(pod, metav1.GetOptions{})
	if err != nil {
		log.Errorf("error : get spark pod %s/%s status error : %v !", ns, pod, err)
		buildErrorResponse(c, err)
		return
	}

	sacommon.NewGinSuccessMessageResponse(podInfo, nil, "success","", c)
}


