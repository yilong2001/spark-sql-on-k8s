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
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"unicode/utf8"

	"github.com/google/go-cloud/blob"
	//"github.com/spf13/cobra"
	log "github.com/sirupsen/logrus"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	goyaml "gopkg.in/yaml.v2"

	clientset "k8s.io/client-go/kubernetes"

	"github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	crdclientset "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned"

	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
)

const bufferSize = 1024
const rootPath = "spark-app-dependencies"

var UploadToPath string = ""
var UploadToEndpoint string = ""
var UploadToRegion string = ""
var Public bool = false
var Override bool = false
var From string = ""

//var Namespace string = "spark-jobs"
const spark_UI_k8s_base = "spark.ui.k8sbase"

/**************************************
add on info to spark application 20201211
**************************************/
func appendExpendInfoOnSparkApplication(namespace string, app *v1beta2.SparkApplication) {
	if (app.Spec.SparkConf == nil) {
		app.Spec.SparkConf = make(map[string]string)
	}
	uiurl, _ := sacommon.GetSparkAppRoutePath4UI(app.Namespace, app.Name)
	app.Spec.SparkConf[spark_UI_k8s_base] = uiurl
	//  "/proxy/spark-ui/"+app.Namespace+"/"+app.Name
}

func createFromYaml(namespace string, yamlFile string, kubeClient clientset.Interface, crdClient crdclientset.Interface) (*v1beta2.SparkApplication, error) {
	var afterApp *v1beta2.SparkApplication
	app, err := loadFromYAML(yamlFile)
	if err != nil {
		log.Errorf("load spark app yaml error : %v", err)
		return nil, fmt.Errorf("failed to read a SparkApplication from : %v", err)
	}

	//log.Infof("\n %v \n", app)

	if afterApp, err = createSparkApplication(namespace, app, kubeClient, crdClient); err != nil {
		log.Errorf("create spark by yaml error : %v", err)
		return nil, fmt.Errorf("failed to create SparkApplication %s: %v", app.Name, err)
	}

	return afterApp, nil
}

func createFromYamlBytes(namespace string, yamlBytes []byte, kubeClient clientset.Interface, crdClient crdclientset.Interface) (*v1beta2.SparkApplication, error) {
	var afterApp *v1beta2.SparkApplication
	app, err := loadFromYAMLBytes(yamlBytes)
	if err != nil {
		log.Errorf("load spark app yaml error : %v", err)
		return nil, fmt.Errorf("failed to read a SparkApplication from : %v", err)
	}

	log.Infof("\n %#v \n", app)

	if afterApp, err = createSparkApplication(namespace, app, kubeClient, crdClient); err != nil {
		log.Errorf("create spark by yaml error : %v", err)
		return nil, fmt.Errorf("failed to create SparkApplication %s: %v", app.Name, err)
	}

	return afterApp, nil
}


func createFromScheduledSparkApplication(namespace string, name string, kubeClient clientset.Interface, crdClient crdclientset.Interface) (*v1beta2.SparkApplication, error) {
	sapp, err := crdClient.SparkoperatorV1beta2().ScheduledSparkApplications(namespace).Get(From, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ScheduledSparkApplication %s: %v", From, err)
	}

	app := &v1beta2.SparkApplication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: v1beta2.SchemeGroupVersion.String(),
					Kind:       reflect.TypeOf(v1beta2.ScheduledSparkApplication{}).Name(),
					Name:       sapp.Name,
					UID:        sapp.UID,
				},
			},
		},
		Spec: *sapp.Spec.Template.DeepCopy(),
	}

	var afterApp *v1beta2.SparkApplication
	if afterApp, err = createSparkApplication(namespace, app, kubeClient, crdClient); err != nil {
		return nil, fmt.Errorf("failed to create SparkApplication %s: %v", app.Name, err)
	}

	return afterApp, nil
}

func createSparkApplication(namespace string, app *v1beta2.SparkApplication, kubeClient clientset.Interface, crdClient crdclientset.Interface) (*v1beta2.SparkApplication, error) {
	v1beta2.SetSparkApplicationDefaults(app)
	if err := validateSpec(app.Spec); err != nil {
		return nil, err
	}

	appendExpendInfoOnSparkApplication(namespace, app)

	if err := handleLocalDependencies(app); err != nil {
		return nil, err
	}

	if hadoopConfDir := os.Getenv("HADOOP_CONF_DIR"); hadoopConfDir != "" {
		fmt.Println("creating a ConfigMap for Hadoop configuration files in HADOOP_CONF_DIR")
		if err := handleHadoopConfiguration(namespace, app, hadoopConfDir, kubeClient); err != nil {
			return nil, err
		}
	}

	var afterApp *v1beta2.SparkApplication
	var err error
	if afterApp, err = crdClient.SparkoperatorV1beta2().SparkApplications(namespace).Create(app); err != nil {
		return nil, err
	}

	log.Infof("SparkApplication \"%s\" created\n", app.Name)

	return afterApp, nil
}

func loadFromYAML(yamlFile string) (*v1beta2.SparkApplication, error) {
	file, err := os.Open(yamlFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewYAMLOrJSONDecoder(file, bufferSize)
	app := &v1beta2.SparkApplication{}
	err = decoder.Decode(app)
	if err != nil {
		log.Errorf("parse from yaml file failed %s : %v", yamlFile, err)
		return nil, err
	}

	return app, nil
}

func loadFromYAMLBytes(yamlBytes []byte) (*v1beta2.SparkApplication, error) {
	app := &v1beta2.SparkApplication{}
	err := goyaml.Unmarshal(yamlBytes, app)
	if err != nil {
		log.Errorf("parse from yaml bytes failed : %v", err)
		return nil, err
	}

	return app, nil
}


func validateSpec(spec v1beta2.SparkApplicationSpec) error {
	if spec.Image == nil && (spec.Driver.Image == nil || spec.Executor.Image == nil) {
		return fmt.Errorf("'spec.driver.image' and 'spec.executor.image' cannot be empty when 'spec.image' " +
			"is not set")
	}

	return nil
}

func handleLocalDependencies(app *v1beta2.SparkApplication) error {
	if app.Spec.MainApplicationFile != nil {
		isMainAppFileLocal, err := isLocalFile(*app.Spec.MainApplicationFile)
		if err != nil {
			return err
		}

		if isMainAppFileLocal {
			uploadedMainFile, err := uploadLocalDependencies(app, []string{*app.Spec.MainApplicationFile})
			if err != nil {
				return fmt.Errorf("failed to upload local main application file: %v", err)
			}
			app.Spec.MainApplicationFile = &uploadedMainFile[0]
		}
	}

	localJars, err := filterLocalFiles(app.Spec.Deps.Jars)
	if err != nil {
		return fmt.Errorf("failed to filter local jars: %v", err)
	}

	if len(localJars) > 0 {
		uploadedJars, err := uploadLocalDependencies(app, localJars)
		if err != nil {
			return fmt.Errorf("failed to upload local jars: %v", err)
		}
		app.Spec.Deps.Jars = uploadedJars
	}

	localFiles, err := filterLocalFiles(app.Spec.Deps.Files)
	if err != nil {
		return fmt.Errorf("failed to filter local files: %v", err)
	}

	if len(localFiles) > 0 {
		uploadedFiles, err := uploadLocalDependencies(app, localFiles)
		if err != nil {
			return fmt.Errorf("failed to upload local files: %v", err)
		}
		app.Spec.Deps.Files = uploadedFiles
	}

	localPyFiles, err := filterLocalFiles(app.Spec.Deps.PyFiles)
	if err != nil {
		return fmt.Errorf("failed to filter local pyfiles: %v", err)
	}

	if len(localPyFiles) > 0 {
		uploadedPyFiles, err := uploadLocalDependencies(app, localPyFiles)
		if err != nil {
			return fmt.Errorf("failed to upload local pyfiles: %v", err)
		}
		app.Spec.Deps.PyFiles = uploadedPyFiles
	}

	return nil
}

func filterLocalFiles(files []string) ([]string, error) {
	var localFiles []string
	for _, file := range files {
		if isLocal, err := isLocalFile(file); err != nil {
			return nil, err
		} else if isLocal {
			localFiles = append(localFiles, file)
		}
	}

	return localFiles, nil
}

func isLocalFile(file string) (bool, error) {
	fileUrl, err := url.Parse(file)
	if err != nil {
		return false, err
	}

	if fileUrl.Scheme == "file" || fileUrl.Scheme == "" {
		return true, nil
	}

	return false, nil
}

type blobHandler interface {
	// TODO: With go-cloud supporting setting ACLs, remove implementations of interface
	setPublicACL(ctx context.Context, bucket string, filePath string) error
}

type uploadHandler struct {
	blob             blobHandler
	blobUploadBucket string
	blobEndpoint     string
	hdpScheme        string
	ctx              context.Context
	b                *blob.Bucket
}

func (uh uploadHandler) uploadToBucket(uploadPath, localFilePath string, override bool) (string, error) {
	fileName := filepath.Base(localFilePath)
	uploadFilePath := filepath.Join(uploadPath, fileName)
	fmt.Printf("uploading file to s3: %s -> %s \n", localFilePath, uploadFilePath)

	// Check if exists by trying to fetch metadata
	reader, err := uh.b.NewRangeReader(uh.ctx, uploadFilePath, 0, 0)
	if err == nil {
		fmt.Errorf("new range reader failed : %s", err)
		reader.Close()
	}
	
	if (blob.IsNotExist(err)) || (err == nil && override) {
		fmt.Printf("uploading local file: %s\n", fileName)

		// Prepare the file for upload.
		data, err := ioutil.ReadFile(localFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %s", err)
		}

		// Open Bucket
		w, err := uh.b.NewWriter(uh.ctx, uploadFilePath, nil)
		if err != nil {
			return "", fmt.Errorf("failed to obtain bucket writer: %s", err)
		}

		// Write data to bucket and close bucket writer
		_, writeErr := w.Write(data)
		if err := w.Close(); err != nil {
			return "", fmt.Errorf("failed to close bucket writer: %s", err)
		}

		// Check if write has been successful
		if writeErr != nil {
			return "", fmt.Errorf("failed to write to bucket: %s", err)
		}

		// Set public ACL if needed
		if Public {
			err := uh.blob.setPublicACL(uh.ctx, uh.blobUploadBucket, uploadFilePath)
			if err != nil {
				return "", err
			}

			endpointURL, err := url.Parse(uh.blobEndpoint)
			if err != nil {
				return "", err
			}
			// Public needs full bucket endpoint
			return fmt.Sprintf("%s://%s/%s/%s",
				endpointURL.Scheme,
				endpointURL.Host,
				uh.blobUploadBucket,
				uploadFilePath), nil
		}
	} else if err == nil {
		fmt.Printf("not uploading file %s as it already exists remotely\n", fileName)
	} else {
		return "", err
	}
	// Return path to file with proper hadoop-connector scheme
	return fmt.Sprintf("%s://%s/%s", uh.hdpScheme, uh.blobUploadBucket, uploadFilePath), nil
}

func uploadLocalDependencies(app *v1beta2.SparkApplication, files []string) ([]string, error) {
	if UploadToPath == "" {
		return nil, fmt.Errorf(
			"unable to upload local dependencies: no upload location specified via --upload-to")
	}

	uploadLocationUrl, err := url.Parse(UploadToPath)
	if err != nil {
		return nil, err
	}
	uploadBucket := uploadLocationUrl.Host

	var uh *uploadHandler
	ctx := context.Background()
	switch uploadLocationUrl.Scheme {
	case "gs":
		uh, err = newGCSBlob(ctx, uploadBucket, UploadToEndpoint, UploadToRegion)
	case "s3":
		uh, err = newS3Blob(ctx, uploadBucket, UploadToEndpoint, UploadToRegion)
	default:
		return nil, fmt.Errorf("unsupported upload location URL scheme: %s", uploadLocationUrl.Scheme)
	}

	// Check if bucket has been successfully setup
	if err != nil {
		return nil, err
	}

	var uploadedFilePaths []string
	uploadPath := filepath.Join(rootPath, app.Namespace, app.Name)
	for _, localFilePath := range files {
		uploadFilePath, err := uh.uploadToBucket(uploadPath, localFilePath, false)
		if err != nil {
			return nil, err
		}

		uploadedFilePaths = append(uploadedFilePaths, uploadFilePath)
	}

	return uploadedFilePaths, nil
}

func handleHadoopConfiguration(
	namespace string,
	app *v1beta2.SparkApplication,
	hadoopConfDir string,
	kubeClientset clientset.Interface) error {
	configMap, err := buildHadoopConfigMap(namespace, app.Name, hadoopConfDir)
	if err != nil {
		return fmt.Errorf("failed to create a ConfigMap for Hadoop configuration files in %s: %v",
			hadoopConfDir, err)
	}

	err = kubeClientset.CoreV1().ConfigMaps(namespace).Delete(configMap.Name, &metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete existing ConfigMap %s: %v", configMap.Name, err)
	}

	if configMap, err = kubeClientset.CoreV1().ConfigMaps(namespace).Create(configMap); err != nil {
		return fmt.Errorf("failed to create ConfigMap %s: %v", configMap.Name, err)
	}

	app.Spec.HadoopConfigMap = &configMap.Name

	return nil
}

func buildHadoopConfigMap(namespace string, appName string, hadoopConfDir string) (*apiv1.ConfigMap, error) {
	info, err := os.Stat(hadoopConfDir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", hadoopConfDir)
	}

	files, err := ioutil.ReadDir(hadoopConfDir)
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Hadoop configuration file found in %s", hadoopConfDir)
	}

	hadoopStringConfigFiles := make(map[string]string)
	hadoopBinaryConfigFiles := make(map[string][]byte)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := ioutil.ReadFile(filepath.Join(hadoopConfDir, file.Name()))
		if err != nil {
			return nil, err
		}

		if utf8.Valid(content) {
			hadoopStringConfigFiles[file.Name()] = string(content)
		} else {
			hadoopBinaryConfigFiles[file.Name()] = content
		}
	}

	configMap := &apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName + "-hadoop-config",
			Namespace: namespace,
		},
		Data:       hadoopStringConfigFiles,
		BinaryData: hadoopBinaryConfigFiles,
	}

	return configMap, nil
}

