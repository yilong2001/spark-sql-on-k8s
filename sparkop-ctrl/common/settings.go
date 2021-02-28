package common

import (
	"fmt"
	//log "github.com/sirupsen/logrus"
	//"github.com/gin-gonic/gin"
	//restful "github.com/emicklei/go-restful"
)

func GetSparkAppUIServiceName(appname string) string {
	return fmt.Sprintf("%s-ui-svc", appname)
}

func GetSparkAppJDBCServiceName(appname string) string {
	return fmt.Sprintf("%s-jdbc-svc", appname)
}

func GetSparkAppRoutePath(namespace, servicename string) (string, error) {
	if (servicename == "" || namespace == "") {
		return "", fmt.Errorf("error : namespace or servicename is empty")
	}

	return fmt.Sprintf("/proxy/spark/%s/%s", namespace, servicename), nil
}

func GetSparkAppRoutePath4UI(namespace, appname string) (string, error) {
	if (appname == "" || namespace == "") {
		return "", fmt.Errorf("error : namespace or appname is empty")
	}

	return fmt.Sprintf("/proxy/spark/%s/%s", namespace, appname), nil
}

func GetSparkAppRoutePath4JDBC(namespace, appname string) (string, error) {
	if (appname == "" || namespace == "") {
		return "", fmt.Errorf("error : namespace or appname is empty")
	}

	return fmt.Sprintf("/proxy/spark-jdbc/%s/%s", namespace, appname), nil
}

