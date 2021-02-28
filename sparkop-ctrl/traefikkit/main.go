package main

import (
	"context"
	//"flag"
	//"fmt"
	"os"
	"os/signal"
	//"strings"
	"syscall"
	//"time"

	"github.com/spf13/pflag"

	//apiv1 "k8s.io/api/core/v1"
	//apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//apitypes "k8s.io/apimachinery/pkg/types"
	//"k8s.io/apimachinery/pkg/util/clock"
	//clientset "k8s.io/client-go/kubernetes"
	log "github.com/sirupsen/logrus"

	//"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"

	//"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	//"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"

	//stan "github.com/nats-io/stan.go"

	myctrl "github.com/spark-sql-on-k8s/sparkop-ctrl/traefikkit/controller"
	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	myutil "github.com/spark-sql-on-k8s/sparkop-ctrl/traefikkit/util"
)

func main() {
	//flag.Parse()
	
	sacommon.InitLogDefault()

	traefikCfg := myutil.NewTraefikCtrlConfig()
	traefikCfg.AddFlags(pflag.CommandLine)
	pflag.Parse()

	traefikCfg.SetLog()

	//traefikCfg.MasterHost = ""
	//traefikCfg.KubeConfig = "/var/lib/rancher/k3s/server/cred/admin.kubeconfig"
	//traefikCfg.SparkAppNamespace = "spark-jobs"
	//traefikCfg.LabelSelectorFilter = ""
	//traefikCfg.ResyncInterval = 0

	//traefikCfg.RouteHostFormat = "spark." + traefikCfg.DomainName
	//traefikCfg.RoutePathPrefixFormat = "/proxy/spark-ui/{namespace}/{name}"

    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)

	stopCh := make(chan struct{}, 1)

	ingressCtrl := myctrl.NewIngressCtrl(traefikCfg)
	defer ingressCtrl.Stop(ctx)

	if err := ingressCtrl.Start(stopCh); err != nil {
		log.Fatal(err)
	}

	go ingressCtrl.Run(ctx)

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
    select {
    case <-signalCh:
    	cancel()
    	close(stopCh)
    case <-stopCh:
    }

    log.Info("Shutting down ... ")
}

