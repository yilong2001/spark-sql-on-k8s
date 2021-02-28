package controller

import (
	"context"
	"encoding/json" 
	//"bytes"
	//"errors"
	//"plugin"
	//"flag"
	"fmt"
	//"os"
	//"os/signal"
	//"strings"
	//"syscall"
	//"time"

	//"github.com/golang/glog"
	//"k8s.io/klog"
	log "github.com/sirupsen/logrus"
	//"github.com/satori/go.uuid"
	//"github.com/gin-gonic/gin"
	//"net/http"

	//"path/filepath"
	//"runtime"
	//"strconv"
	//"strings"

	//apiv1 "k8s.io/api/core/v1"
	//apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	//"k8s.io/apimachinery/pkg/util/clock"
	//"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
	//"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/client-go/tools/record"

	crdclientset "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"

	stan "github.com/nats-io/stan.go"

	sacommon "github.com/spark-sql-on-k8s/sparkop-ctrl/common"
	myutil "github.com/spark-sql-on-k8s/sparkop-ctrl/traefikkit/util"
)

const natsClientID  = "sparkapp-traefik-ctrl"

type  IngressCtrl struct {
	cfg           *myutil.TraefikCtrlConfig
	kubeClient    clientset.Interface
	crdClient     crdclientset.Interface
	natsStreamClient *sacommon.NatsStreamingClient
}

func NewIngressCtrl(traefikCfg  *myutil.TraefikCtrlConfig) *IngressCtrl {
	config, err := sacommon.BuildK8sRestConfig(traefikCfg.MasterHost, traefikCfg.KubeConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("build rest config ok! ")
	//log.Info("build rest config ok! ")

	kubeclient, err := clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("build kube clientset config ok! ")

	crdclient, err := crdclientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("build traefik crd clientset config ok! ")

	return &IngressCtrl{
		cfg: traefikCfg,
		kubeClient: kubeclient,
		crdClient: crdclient,
		natsStreamClient: sacommon.NewNatsStreamingClient(),
	}
}

func (r *IngressCtrl) Start(stopCh <-chan struct{}) error {
	r.natsStreamClient.Start(stopCh, r.cfg.NatsStreamUrl, natsClientID)
	return nil
}

func (r *IngressCtrl) Stop(ctx context.Context) {
    r.natsStreamClient.Stop(ctx)
}

func (r *IngressCtrl) getRoutePath(namespace string, ss *sacommon.SparkService) string {
    path, _ := sacommon.GetSparkAppRoutePath(namespace, ss.ServiceName)
    return path
}

func (r *IngressCtrl) getRouteHost() string {
    //host := strings.ReplaceAll(r.cfg.RouteHostFormat, "{name}", msg.Name)
    //host = strings.ReplaceAll(host, "{namespace}", msg.Namespace)  
    return "spark." + r.cfg.DomainName
}

func (r *IngressCtrl) Run(ctx context.Context) {
    dataCh := make(chan string, 1024)

    sub, err := r.natsStreamClient.GetConn().Subscribe(sacommon.NatsChannelTraefik, func(msg *stan.Msg) {
		msg.Ack()
		dataCh <- string(msg.Data)

		// Print the value and whether it was redelivered.
		log.Infof("seq = %d [redelivered = %v]\n", msg.Sequence, msg.Redelivered)
	}, stan.DurableName(sacommon.NatsDurableID), stan.MaxInflight(1), stan.SetManualAckMode())
	if err != nil {
		log.Fatal(err)
	}
	
	defer sub.Close()
	defer close(dataCh)

	for {
		select {
			case data := <- dataCh:
				var msg sacommon.SparkAppCreatedMsg
				err := json. Unmarshal([]byte(data) , &msg)
				if err != nil {
					log.Errorf(" %s unmarshal error : %v ", data, err)
				} else {
					r.createTraefikIngressRouteResource(msg)
				}
			case <-ctx.Done():
				return
		}
	}    
}

func (r *IngressCtrl) createTraefikIngressRouteResource(
	msg sacommon.SparkAppCreatedMsg) error {
    //ingressRoute := r.buildTraefikIngressRoute(msg)
    //_, err := r.crdClient.TraefikV1alpha1().IngressRoutes(msg.Namespace).Create(context.TODO(), ingressRoute, metav1.CreateOptions{})
    var err error = nil

    for _, svc := range msg.ExtraServices {
    	err = r.buildTraefikIngressRoute(msg.Namespace, msg.Name, msg.UID, svc)
	    if err != nil {
    		log.Errorf("create traefik ingress route resource error : %v", err)
        	return err
    	}
    }

    log.Infof("create traefik ingress route resource ok: %s-%s-%s", msg.Namespace, msg.Name, msg.UID)

    return nil
}

/*
func (r *IngressCtrl) getServicesInRouter(namespace string, msg sacommon.SparkAppCreatedMsg) []v1alpha1.Service {
	services := make([]v1alpha1.Service, 0)
	if (msg.ExtraServices == nil) {
		return services
	}

	for _, svc := range msg.ExtraServices {
		services = append(services, v1alpha1.LoadBalancerSpec{
               Name: svc.ServiceName,
               Port: svc.ServicePort,
               Kind: "Service",
               Namespace: namespace,
           })
	}

	return services
}

func (r *IngressCtrl) BuildTraefikIngressRoute(msg sacommon.SparkAppCreatedMsg, 
	namespace, name, uid, host, pathpre string) *v1alpha1.IngressRoute {
    ingress := &v1alpha1.IngressRoute {
		ObjectMeta: metav1.ObjectMeta {
			Name:            name,
			Namespace:       namespace,
			//Labels:          getResourceLabels(app),
			OwnerReferences: []metav1.OwnerReference{*r.getSparkAppOwnerReference(name, uid)},
		},
		Spec: v1alpha1.IngressRouteSpec {
		    Routes: []v1alpha1.Route {{
		       Match: fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", host, pathpre),
		       Kind: "Rule",
		       Services: r.getServicesInRouter(namespace, msg),
		    }},
		    EntryPoints: []string {"web"},
		 },
	}

    return ingress
}

func (r *IngressCtrl) buildTraefikIngressRoute(msg sacommon.SparkAppCreatedMsg) *v1alpha1.IngressRoute {
    routes := make([]v1alpha1.Route, 0)

	for _, svc := range msg.ExtraServices {
		route := v1alpha1.Route {
	        Match: fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", 
	       		r.getRouteHost(), svc.RoutePath),
	        Kind: "Rule",
       		Services: []v1alpha1.Service{{
	           LoadBalancerSpec: v1alpha1.LoadBalancerSpec{
	                Name: svc.ServiceName,
           			Port: svc.ServicePort,
           			Kind: "Service",
           			Namespace: msg.Namespace,
	           },
	       }},
	    }
		routes = append(routes, route)
	}

    ingress := &v1alpha1.IngressRoute {
		ObjectMeta: metav1.ObjectMeta {
			Name:            msg.Name,
			Namespace:       msg.Namespace,
			//Labels:          getResourceLabels(app),
			OwnerReferences: []metav1.OwnerReference{*r.getSparkAppOwnerReference(msg.Name, msg.UID)},
		},
		Spec: v1alpha1.IngressRouteSpec {
		    Routes: routes,
		    EntryPoints: []string {"web"},
		},
	}

    return ingress
}
*/

func (r *IngressCtrl) buildTraefikIngressRoute(namespace, name string,
	UID string, svc *sacommon.SparkService) error {
	var err error = nil

	if svc.ServiceType == "web" {
		route := v1alpha1.Route {
	        Match: fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", 
	       		r.getRouteHost(), svc.RoutePath),
	        Kind: "Rule",
	   		Services: []v1alpha1.Service{{
	           LoadBalancerSpec: v1alpha1.LoadBalancerSpec{
	                Name: svc.ServiceName,
	       			Port: svc.ServicePort,
	       			Kind: "Service",
	       			Namespace: namespace,
	           },
	       }},
	    }
	    ingress := &v1alpha1.IngressRoute {
			ObjectMeta: metav1.ObjectMeta {
				Name:            svc.ServiceName,
				Namespace:       namespace,
				OwnerReferences: []metav1.OwnerReference{*r.getSparkAppOwnerReference(name, UID)},
			},
			Spec: v1alpha1.IngressRouteSpec {
			    Routes: []v1alpha1.Route{route},
			    EntryPoints: []string {svc.EntryPoint},
			},
		}
		_, err = r.crdClient.TraefikV1alpha1().IngressRoutes(namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
	} else if svc.ServiceType == "tcp" {
		route := v1alpha1.RouteTCP {
	        Match: "HostSNI(`*`)",
	   		Services: []v1alpha1.ServiceTCP{{
	   			Name: svc.ServiceName,
	   			Namespace: namespace,
	   			Port: svc.ServicePort,
	   		}},
	    }
	    ingress := &v1alpha1.IngressRouteTCP {
			ObjectMeta: metav1.ObjectMeta {
				Name:            svc.ServiceName,
				Namespace:       namespace,
				OwnerReferences: []metav1.OwnerReference{*r.getSparkAppOwnerReference(name, UID)},
			},
			Spec: v1alpha1.IngressRouteTCPSpec {
			    Routes: []v1alpha1.RouteTCP{route},
			    EntryPoints: []string {svc.EntryPoint},
			},
		}
		_, err = r.crdClient.TraefikV1alpha1().IngressRouteTCPs(namespace).Create(context.TODO(), ingress, metav1.CreateOptions{})
	} else {
		// TODO:
	}
	
    return err
}

func (r *IngressCtrl) getSparkAppOwnerReference(name, uid string) *metav1.OwnerReference {
	return &metav1.OwnerReference {
		APIVersion: "sparkoperator.k8s.io/v1beta2",
		Kind:       "SparkApplication",
		Name:       name,
		UID:        apitypes.UID(uid),
	}
}

