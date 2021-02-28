package controller

import (
	"context"
	//"bytes"
	//"errors"
	//"plugin"
	//"flag"
	"fmt"
	"strconv"
	//"encoding/json" 
	//"os"
	//"net/url"
	//"net/http"
	//"path/filepath"
	//"os/signal"
	//"strings"
	"reflect"
	//"syscall"
	//"time"

	//"github.com/golang/glog"
	//"k8s.io/klog"
	log "github.com/sirupsen/logrus"
	//"github.com/satori/go.uuid"
	//"github.com/gin-gonic/gin"

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
	"k8s.io/apimachinery/pkg/util/intstr"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/config"

	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
)

const (
	sparkUIPortConfigurationKey       = "spark.ui.port"
	
	sparkJdbcPortConfigurationKey     = "spark.thrift.jdbc.bind.port"

	defaultSparkWebUIPort       int32 = 4040

	defaultThriftJDBCPort int32 = 10000
)

func getDriverUITargetPort(app *v1beta2.SparkApplication) (int32, error) {
	portStr, ok := app.Spec.SparkConf[sparkUIPortConfigurationKey]
	if ok {
		port, err := strconv.Atoi(portStr)
		return int32(port), err
	}
	return defaultSparkWebUIPort, nil
}

func getDriverUIServicePort(app *v1beta2.SparkApplication) (int32, error) {
	if app.Spec.SparkUIOptions == nil {
		return getDriverUITargetPort(app)
	}

	port := app.Spec.SparkUIOptions.ServicePort
	if port != nil {
		return *port, nil
	}
	return defaultSparkWebUIPort, nil
}

func getDriverUIService(app *v1beta2.SparkApplication) (*sacommon.SparkService, error) {
	targetPort, err := getDriverUITargetPort(app)
	if err != nil {
		return nil, err
	}

	servicePort, err := getDriverUIServicePort(app)
	if err != nil {
		return nil, err
	}

	routepath, _ := sacommon.GetSparkAppRoutePath4UI(app.Namespace, app.Name)
	return &sacommon.SparkService {
		ServiceName: fmt.Sprintf("%s-ui-svc", app.Name),
		ServiceType: "web",
		EntryPoint: "web",
		ServicePort: servicePort,
		TargetPort: int32(targetPort),
		ServiceIP: "",
		RoutePath:  routepath,
	}, nil
}

func getJDBCServicePort(app *v1beta2.SparkApplication) (int32, error) {
	if app.Spec.SparkConf == nil {
		return 0, fmt.Errorf("not spark sql application")
	}

	portStr,ok := app.Spec.SparkConf[sparkJdbcPortConfigurationKey]
	if !ok {
		return 0, fmt.Errorf("not spark sql application")
	}


	port, err := strconv.Atoi(portStr)
	return int32(port), err
}

func getJDBCTargetPort(app *v1beta2.SparkApplication) (int32, error) {
	if app.Spec.SparkConf == nil {
		return 0, fmt.Errorf("not spark sql application")
	}
	
	portStr,ok := app.Spec.SparkConf[sparkJdbcPortConfigurationKey]
	if !ok {
		return 0, fmt.Errorf("not spark sql application")
	}

	port, err := strconv.Atoi(portStr)
	return int32(port), err
}

func getResourceLabels(app *v1beta2.SparkApplication) map[string]string {
	labels := map[string]string{config.SparkAppNameLabel: app.Name}
	if app.Status.SubmissionID != "" {
		labels[config.SubmissionIDLabel] = app.Status.SubmissionID
	}
	return labels
}

func getOwnerReference(app *v1beta2.SparkApplication) *metav1.OwnerReference {
	return &metav1.OwnerReference{
		APIVersion: v1beta2.SchemeGroupVersion.String(),
		Kind:       reflect.TypeOf(v1beta2.SparkApplication{}).Name(),
		Name:       app.Name,
		UID:        app.UID,
	}
}

func (r *SparkAppCtrl) createSparkJDBCService(app *v1beta2.SparkApplication) (*sacommon.SparkService, error) {
	port, err := getJDBCServicePort(app)
	if err != nil {
		return nil, err
	}
	tPort, err := getJDBCTargetPort(app)
	if err != nil {
		return nil, err
	}
	service := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            sacommon.GetSparkAppJDBCServiceName(app.Name),
			Namespace:       app.Namespace,
			Labels:          getResourceLabels(app),
			OwnerReferences: []metav1.OwnerReference{*getOwnerReference(app)},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name: "spark-driver-jdbc-port",
					Port: port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: tPort,
					},
				},
			},
			Selector: map[string]string{
				config.SparkAppNameLabel: app.Name,
				config.SparkRoleLabel:    config.SparkDriverRole,
			},
			Type: apiv1.ServiceTypeClusterIP,
		},
	}

	log.Infof("Creating a service %s for the Spark thrift jdbc for application %s", service.Name, app.Name)
	service, err = r.kubeClient.CoreV1().Services(app.Namespace).Create(service)
	if err != nil {
		return nil, err
	}

	routepath, _ := sacommon.GetSparkAppRoutePath4JDBC(app.Namespace, app.Name)
	return &sacommon.SparkService {
		ServiceName: service.Name,
		ServiceType: "tcp",
		EntryPoint: "hive2",
		ServicePort: service.Spec.Ports[0].Port,
		TargetPort:  service.Spec.Ports[0].TargetPort.IntVal,
		ServiceIP:   service.Spec.ClusterIP,
		RoutePath:  routepath,
	}, nil
}

/*****************************************************

*****************************************************/
func (r *SparkAppCtrl) getSparkService(service *apiv1.Service) *sacommon.SparkService {
	return &sacommon.SparkService {
		ServiceName: service.Name,
		ServiceType: "tcp",
		EntryPoint: "hive2",
		ServicePort: service.Spec.Ports[0].Port,
		TargetPort:  service.Spec.Ports[0].TargetPort.IntVal,
		ServiceIP:   service.Spec.ClusterIP,
		// RoutePath:  routepath,
	}
}

func (r *SparkAppCtrl) getSparkServiceStatus(ctx context.Context, 
	namespace, servicename string) (*apiv1.Service, error) {
	return r.kubeClient.CoreV1().Services(namespace).Get(servicename, metav1.GetOptions{})
}

