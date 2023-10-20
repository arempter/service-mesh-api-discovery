package k8s

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/exp/slog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sCollector is implementing data lookup for Kubernetes objects using Endpoints sharedInformer
type K8sCollector interface {
	// returns endpoint name for given IP address
	LookupFor(ipAddr string) string
}

type k8sCollector struct {
	endpointsInf cache.SharedIndexInformer
	endpintsIPs  map[string]string // IP address to name resolution structure
}

func NewK8sCollector(clientset kubernetes.Interface) *k8sCollector {
	factory := informers.NewSharedInformerFactory(clientset, 1*time.Hour)

	return &k8sCollector{
		endpointsInf: factory.Core().V1().Endpoints().Informer(),
		endpintsIPs:  make(map[string]string),
	}
}

// Clientset returns Clientset either for inCluster or local rest config or panics
func Clientset() *kubernetes.Clientset {
	var (
		config *rest.Config
		err    error
	)

	config, err = rest.InClusterConfig()
	if err == nil {
		slog.Info("found InClusterConfig")
	} else {
		localConfig := os.Getenv("KUBECONFIG")
		if localConfig == "" {
			panic("unable to obtain KUBECONIF")
		}
		config, err = clientcmd.BuildConfigFromFlags("", localConfig)
		if err != nil {
			slog.Error("failed to create k8s config, err: %s", "err", err.Error())
			panic("failed to create k8s config")
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return clientset
}

// registerHandlers only implements addFunc so updates or deltes will not be reflected. It is enough for this demo flow
func (c *k8sCollector) registerHandlers() {
	c.endpointsInf.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			e := obj.(*v1.Endpoints)
			slog.Info("endpoint added", "namespace", e.GetNamespace(), "endpointName", e.GetName())
			for _, s := range e.Subsets {
				for _, ipAddr := range s.Addresses {
					slog.Info("populating endpoints", "ipAddr", ipAddr.IP, "endpointName", e.GetName())
					c.endpintsIPs[ipAddr.IP] = e.GetName()
				}
			}
		},
		// not implemented
		UpdateFunc: func(oldObj, newObj interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})
}

// Run starts endpoints informer
func (c *k8sCollector) Run(stop chan struct{}) {
	slog.Info("starting k8s endpoints collector")
	c.registerHandlers()

	go c.endpointsInf.Run(stop)
	if !cache.WaitForCacheSync(stop, c.endpointsInf.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	} else {
		slog.Info("endpoints collector informer cache in sync")
	}

	<-stop
	slog.Info("shutting down k8s collector")
}

// LookupFor return endpoint (resource) name for given IP address
func (c *k8sCollector) LookupFor(ipAddr string) string {
	entry, found := c.endpintsIPs[ipAddr]
	if found {
		return entry
	}
	return ""
}
