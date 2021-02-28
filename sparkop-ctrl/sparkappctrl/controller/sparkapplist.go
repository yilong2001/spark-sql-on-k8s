/*
Copyright 2017 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"time"
	"k8s.io/apimachinery/pkg/util/duration"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"
)

type SparkAppInfo struct {
	Namespace string
	Name      string
	State     string
	SubmissionAge   string
	TerminationAge  string
	DriverUrl  string
	JdbcSvc  string
}

func getSinceTime(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return "N.A."
	}

	return duration.ShortHumanDuration(time.Since(timestamp.Time))
}

func formatNotAvailable(info string) string {
	if info == "" {
		return "N.A."
	}
	return info
}

func listApplication(namespace string, crdClientset crdclientset.Interface) (*v1beta2.SparkApplicationList, error) {
	applist, err := crdClientset.SparkoperatorV1beta2().SparkApplications(namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("error : spark application list %s, %v",namespace,err)
		return nil, err
	}

	return applist, err
}

