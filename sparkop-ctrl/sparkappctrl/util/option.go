package util

import (
	"github.com/spf13/pflag"
	"github.com/sirupsen/logrus"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	
	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
)

type SparkCtrlConfig struct {
	LogLevel    string
	MasterHost  string
	KubeConfig  string
	ResyncInterval    int
	SparkAppNamespace   string
	LabelSelectorFilter  string
	TmpLocalDir      string

	WebBindHost    string
	WebBindPort    int

	MaxUploadFileSize int

	S3UploadDir string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string

	NatsStreamUrl    string
	DomainName  string
	UIEnableHttps bool

	MetaDb  sacommon.DBConnConfig
}

func NewSparkCtrlConfig() *SparkCtrlConfig {
	return &SparkCtrlConfig{}
}

//AddFlags config
func (a *SparkCtrlConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&a.LogLevel, "log-level", "0", " log level")
	fs.StringVar(&a.MasterHost, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&a.KubeConfig, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")

	fs.IntVar(&a.ResyncInterval, "resync-interval", 30, "Informer resync interval in seconds.")
	fs.StringVar(&a.SparkAppNamespace, "namespace", apiv1.NamespaceAll, "The Kubernetes namespace to manage. Will manage custom resource objects of the managed CRD types for the whole cluster if unset.")
	fs.StringVar(&a.LabelSelectorFilter, "label-selector-filter", "", "A comma-separated list of key=value, or key labels to filter resources during watch and list based on the specified labels.")
	fs.StringVar(&a.TmpLocalDir, "tmp-local-dir", "./tmp", "yaml tmp local dir")

	fs.StringVar(&a.WebBindHost, "web-bind-host", "", "web  srvice bind hostname/ip")
	fs.IntVar(&a.WebBindPort, "web-bind-port", 8085, "web service bind port")

	fs.IntVar(&a.MaxUploadFileSize, "max-upload-file-size", 1024*1024, "max upload file size")

	fs.StringVar(&a.S3UploadDir, "s3-upload-dir", "", "s3 public uploaded file path")
	fs.StringVar(&a.S3Endpoint, "s3-endpoint", "", "s3 endpoint")
	fs.StringVar(&a.S3AccessKey, "s3-accesskey", "", "s3 access key")
	fs.StringVar(&a.S3SecretKey, "s3-secretkey", "", "s3 credit key")

	fs.StringVar(&a.NatsStreamUrl, "nats-stream-url", "", "nats streaming url")

	fs.StringVar(&a.DomainName, "domain-name", "mydomain.io", "http url domain name")
	fs.BoolVar(&a.UIEnableHttps, "ui-enable-https", false, "driver ui enable https")

	fs.StringVar(&a.MetaDb.Type, "metadb-type", "MYSQL", "sql metadb type, default is mysql")
	fs.StringVar(&a.MetaDb.Host, "metadb-host", "", "sql metadb host")
	fs.StringVar(&a.MetaDb.Port, "metadb-port", "3306", "sql metadb port, default is 3306")
	fs.StringVar(&a.MetaDb.User, "metadb-user", "", "sql metadb user")
	fs.StringVar(&a.MetaDb.Password, "metadb-pw", "", "sql metadb password")
	fs.StringVar(&a.MetaDb.Dbname, "metadb-dbname", "hive", "sql metadb db name, default is hive")
}

//SetLog 设置log
func (a *SparkCtrlConfig) SetLog() {
	level, err := logrus.ParseLevel(a.LogLevel)
	if err != nil {
		fmt.Println("set log level error." + err.Error())
		return
	}
	logrus.SetLevel(level)
}


