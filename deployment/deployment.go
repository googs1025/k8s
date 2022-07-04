package deployment

import (
	"context"
	"sync"
	"time"

	//_ "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	//_ "k8s.io/client-go/applyconfigurations/apps/v1"
	//_ "k8s.io/client-go/applyconfigurations/meta/v1"

	"github.com/forbearing/k8s/typed"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type Handler struct {
	kubeconfig string
	namespace  string

	ctx                context.Context
	config             *rest.Config
	restClient         *rest.RESTClient
	clientset          *kubernetes.Clientset
	dynamicClient      dynamic.Interface
	discoveryClient    *discovery.DiscoveryClient
	discoveryInterface discovery.DiscoveryInterface
	informerFactory    informers.SharedInformerFactory
	informer           cache.SharedIndexInformer

	Options *typed.HandlerOptions

	sync.Mutex
}

//// Discovery retrieves the DiscoveryClient
//func (c *Clientset) Discovery() discovery.DiscoveryInterface {
//    if c == nil {
//        return nil
//    }
//    return c.DiscoveryClient
//}
// clientset 调用 Discovery 方法可以得到一个 discovery.DiscoveryInterface
// discovery.DiscoveryClient 其实就是 discovery.DiscoveryInterface 的一个实现

// New returns a deployment handler from kubeconfig or in-cluster config
func New(ctx context.Context, namespace, kubeconfig string) (handler *Handler, err error) {
	var (
		config             *rest.Config
		restClient         *rest.RESTClient
		clientset          *kubernetes.Clientset
		dynamicClient      dynamic.Interface
		discoveryClient    *discovery.DiscoveryClient
		discoveryInterface discovery.DiscoveryInterface
		informerFactory    informers.SharedInformerFactory
	)
	handler = &Handler{}

	// create rest config
	if len(kubeconfig) != 0 {
		// use the current context in kubeconfig
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		// create the in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	// setup APIPath, GroupVersion and NegotiatedSerializer before initializing a RESTClient
	config.APIPath = "api"
	config.GroupVersion = &appsv1.SchemeGroupVersion
	//config.GroupVersion = &schema.GroupVersion{Group: "apps", Version: "v1"}
	config.NegotiatedSerializer = scheme.Codecs
	//config.UserAgent = rest.DefaultKubernetesUserAgent()

	// create a RESTClient for the given config
	restClient, err = rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	// create a Clientset for the given config
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	// create a dynamic client for the given config
	dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	// create a DiscoveryClient for the given config
	discoveryClient, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	// create a sharedInformerFactory for all namespaces.
	informerFactory = informers.NewSharedInformerFactory(clientset, time.Minute)
	//discoveryClient = clientset.DiscoveryClient
	//discoveryInterface = clientset.Discovery()

	// default namespace is meatv1.NamespaceDefault ("default")
	if len(namespace) == 0 {
		namespace = metav1.NamespaceDefault
	}
	handler.kubeconfig = kubeconfig
	handler.namespace = namespace
	handler.ctx = ctx
	handler.config = config
	handler.restClient = restClient
	handler.clientset = clientset
	handler.dynamicClient = dynamicClient
	handler.discoveryClient = discoveryClient
	handler.informerFactory = informerFactory
	handler.informer = informerFactory.Apps().V1().Deployments().Informer()
	//handler.discoveryInterface = discoveryInterface
	_ = discoveryInterface

	handler.Options = &typed.HandlerOptions{}

	return handler, nil
}
func (h *Handler) Namespace() string {
	return h.namespace
}
func (in *Handler) DeepCopy() *Handler {
	out := new(Handler)

	// 值拷贝即是深拷贝
	out.kubeconfig = in.kubeconfig
	out.namespace = in.namespace

	// 和几个字段都是共用的, 不需要深拷贝
	out.ctx = in.ctx
	out.config = in.config
	out.restClient = in.restClient
	out.clientset = in.clientset
	out.dynamicClient = in.dynamicClient
	out.discoveryClient = in.discoveryClient
	out.informerFactory = in.informerFactory
	out.informer = in.informer

	out.Options = &typed.HandlerOptions{}
	out.Options.ListOptions = *in.Options.ListOptions.DeepCopy()
	out.Options.GetOptions = *in.Options.GetOptions.DeepCopy()
	out.Options.CreateOptions = *in.Options.CreateOptions.DeepCopy()
	out.Options.UpdateOptions = *in.Options.UpdateOptions.DeepCopy()
	out.Options.PatchOptions = *in.Options.PatchOptions.DeepCopy()
	out.Options.ApplyOptions = *in.Options.ApplyOptions.DeepCopy()

	// 锁 sync.Mutex 不需要拷贝, 也不能拷贝. 拷贝 sync.Mutex 会直接 panic

	//fmt.Printf("%#v\n", oldHandler)
	//fmt.Println()
	//fmt.Printf("%#v\n", out)
	//select {}

	return out
}
func (h *Handler) setNamespace(namespace string) {
	h.Lock()
	h.Unlock()
	h.namespace = namespace
}
func (h *Handler) WithNamespace(namespace string) *Handler {
	handler := h.DeepCopy()
	handler.setNamespace(namespace)
	return handler
}
func (h *Handler) WithDryRun() *Handler {
	handler := h.DeepCopy()
	handler.Options.CreateOptions.DryRun = []string{metav1.DryRunAll}
	handler.Options.UpdateOptions.DryRun = []string{metav1.DryRunAll}
	handler.Options.DeleteOptions.DryRun = []string{metav1.DryRunAll}
	handler.Options.PatchOptions.DryRun = []string{metav1.DryRunAll}
	handler.Options.ApplyOptions.DryRun = []string{metav1.DryRunAll}
	return handler
}
func (h *Handler) SetTimeout(timeout int64) {
	h.Lock()
	defer h.Unlock()
	h.Options.ListOptions.TimeoutSeconds = &timeout
}
func (h *Handler) SetLimit(limit int64) {
	h.Lock()
	defer h.Unlock()
	h.Options.ListOptions.Limit = limit
}
func (h *Handler) SetForceDelete(force bool) {
	h.Lock()
	defer h.Unlock()
	if force {
		gracePeriodSeconds := int64(0)
		h.Options.DeleteOptions.GracePeriodSeconds = &gracePeriodSeconds
	} else {
		h.Options.DeleteOptions = metav1.DeleteOptions{}
	}
}
