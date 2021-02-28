module github.com/spark-sql-on-k8s/sparkop-ctrl/sparkappctrl

go 1.13

require (
	cloud.google.com/go v0.38.0
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20201203164634-38cfc22dd87e
	github.com/aws/aws-sdk-go v1.29.11
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20200421181703-e76ad31c14f6 // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.2.0 // indirect
	github.com/go-redis/redis v6.15.9+incompatible // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.4 // indirect
	github.com/google/go-cloud v0.1.1
	github.com/google/uuid v1.1.2
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/kataras/iris/v12 v12.1.8
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/minio/minio-go/v7 v7.0.6

	//github.com/micro/go-micro v3.0.2
	github.com/nats-io/stan.go v0.7.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.3.0
	github.com/prometheus/client_model v0.1.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spark-sql-on-k8s/pkg v0.0.0-00010101000000-000000000000
	github.com/spark-sql-on-k8s/sparkop-ctrl/common v0.0.0-00010101000000-000000000000
	github.com/spark-sql-on-k8s/sparkop-ctrl/plugin v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	//github.com/sirupsen/logrus v1.4.2

	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.6 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.29.1 // indirect
	gopkg.in/yaml.v2 v2.2.8
	gorm.io/driver/mysql v1.0.4 // indirect
	honnef.co/go/tools v0.0.1-2020.1.4 // indirect
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v0.18.2
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.16.2
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66 // indirect
	rsc.io/quote/v3 v3.1.0 // indirect
	volcano.sh/volcano v0.4.0
)

replace (

	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
	github.com/gorilla/mux => github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
	github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595
	github.com/mailgun/multibuf => github.com/containous/multibuf v0.0.0-20190809014333-8b6c9a7e6bba
	github.com/spark-sql-on-k8s/pkg => ../pkg
	github.com/spark-sql-on-k8s/sparkop-ctrl/common => ../common
	github.com/spark-sql-on-k8s/sparkop-ctrl/plugin => ../plugin
	//cloud.google.com/go => ./vendors/google-cloud-go
	golang.org/x/sys => /root/go/pkg/mod/golang.org/x/sys@v0.0.0-20200323222414-85ca7c5b95cd

	google.golang.org/api => google.golang.org/api v0.14.0

	google.golang.org/grpc => google.golang.org/grpc v1.26.0
	//github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0

	k8s.io/api => k8s.io/api v0.16.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.6
	k8s.io/apiserver => k8s.io/apiserver v0.16.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.6
	k8s.io/client-go => k8s.io/client-go v0.16.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.6
	k8s.io/code-generator => k8s.io/code-generator v0.16.6
	k8s.io/component-base => k8s.io/component-base v0.16.6
	k8s.io/cri-api => k8s.io/cri-api v0.16.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.6
	k8s.io/kubectl => k8s.io/kubectl v0.16.6
	k8s.io/kubelet => k8s.io/kubelet v0.16.6
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.6
	k8s.io/metrics => k8s.io/metrics v0.16.6
	k8s.io/node-api => k8s.io/node-api v0.16.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.6
	k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.16.6
	k8s.io/sample-controller => k8s.io/sample-controller v0.16.6
)
