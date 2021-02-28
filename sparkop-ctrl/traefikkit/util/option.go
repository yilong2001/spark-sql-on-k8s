package util

import (
	"github.com/spf13/pflag"
	"github.com/sirupsen/logrus"
	"fmt"
	apiv1 "k8s.io/api/core/v1"
)

type TraefikCtrlConfig struct {
	LogLevel    string
	MasterHost  string
	KubeConfig  string
	ResyncInterval    int
	SparkAppNamespace   string
	LabelSelectorFilter  string

	WebBindHost    string
	WebBindPort    int

	MaxUploadFileSize int

	EnableIngressRoute    bool      
	//RouteHostFormat       string  
	//RoutePathPrefixFormat string

	NatsStreamUrl    string
	DomainName  string
	UIEnableHttps bool
}

func NewTraefikCtrlConfig() *TraefikCtrlConfig {
	return &TraefikCtrlConfig{}
}

//AddFlags config
func (a *TraefikCtrlConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&a.LogLevel, "log-level", "0", " log level")
	fs.StringVar(&a.MasterHost, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	fs.StringVar(&a.KubeConfig, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")

	fs.IntVar(&a.ResyncInterval, "resync-interval", 30, "Informer resync interval in seconds.")
	fs.StringVar(&a.SparkAppNamespace, "namespace", apiv1.NamespaceAll, "The Kubernetes namespace to manage. Will manage custom resource objects of the managed CRD types for the whole cluster if unset.")
	fs.StringVar(&a.LabelSelectorFilter, "label-selector-filter", "", "A comma-separated list of key=value, or key labels to filter resources during watch and list based on the specified labels.")

	fs.StringVar(&a.WebBindHost, "web-bind-host", "", "web  srvice bind hostname/ip")
	fs.IntVar(&a.WebBindPort, "web-bind-port", 8095, "web service bind port")

	fs.IntVar(&a.MaxUploadFileSize, "max-upload-file-size", 1024*1024, "max upload file size")

	fs.BoolVar(&a.EnableIngressRoute, "enable-ingress-route", true, "auto create traefik ingress route")
	//fs.StringVar(&a.RouteHostFormat, "route-host-format", "", "host format for route ")
	//fs.StringVar(&a.RoutePathPrefixFormat, "route-path-pre-format", "", "path prefix format for route")

	fs.StringVar(&a.NatsStreamUrl, "nats-stream-url", "", "nats streaming url")
	fs.StringVar(&a.DomainName, "domain-name", "mydomain.io", "http url domain name")
	fs.BoolVar(&a.UIEnableHttps, "ui-enable-https", false, "driver ui enable https")
}

//SetLog 设置log
func (a *TraefikCtrlConfig) SetLog() {
	level, err := logrus.ParseLevel(a.LogLevel)
	if err != nil {
		fmt.Println("set log level error." + err.Error())
		return
	}
	logrus.SetLevel(level)
}


