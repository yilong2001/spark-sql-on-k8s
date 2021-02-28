package controller

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
)

func deleteSparkApplication(namespace, name string, crdClientset crdclientset.Interface) error {
	err := crdClientset.SparkoperatorV1beta2().SparkApplications(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("error : spark application delete %s/%s, %v",namespace,name,err)
		return err
	}

	log.Infof("SparkApplication \"%s\" deleted\n", name)

	return nil
}

