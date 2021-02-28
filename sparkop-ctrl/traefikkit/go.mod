module github.com/spark-sql-on-k8s/sparkop-ctrl/traefikkit

go 1.15

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/googleapis/gnostic v0.4.1
	github.com/nats-io/stan.go v0.7.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spark-sql-on-k8s/sparkop-ctrl/common v0.0.0-00010101000000-000000000000
	github.com/spf13/pflag v1.0.5
	github.com/traefik/traefik/v2 v2.3.4
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
)

// Docker v19.03.6
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203

// Containous forks
replace (
	github.com/spark-sql-on-k8s/sparkop-ctrl/common => ../common
	github.com/spark-sql-on-k8s/sparkop-ctrl/plugin => ../plugin

	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
	github.com/gorilla/mux => github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
	github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595
	github.com/mailgun/multibuf => github.com/containous/multibuf v0.0.0-20190809014333-8b6c9a7e6bba
)
