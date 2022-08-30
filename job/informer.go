package job

import (
	"time"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	informersbatch "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/informers/internalinterfaces"
	listersbatch "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
)

// SetInformerFactoryResyncPeriod will set informer resync period.
func (h *Handler) SetInformerFactoryResyncPeriod(resyncPeriod time.Duration) {
	h.l.Lock()
	defer h.l.Unlock()
	h.resyncPeriod = resyncPeriod
	if len(h.informerScope) == 0 {
		h.informerScope = metav1.NamespaceAll
	}
	h.informerFactory = informers.NewSharedInformerFactoryWithOptions(
		h.clientset, h.resyncPeriod,
		informers.WithNamespace(h.informerScope),
		informers.WithTweakListOptions(h.tweakListOptions))
}

// SetInformerFactoryNamespace limit the scope of informer list-and-watch k8s resource.
// informer list-and-watch all namespace k8s resource by default.
func (h *Handler) SetInformerFactoryNamespace(namespace string) {
	h.l.Lock()
	defer h.l.Unlock()
	h.informerScope = namespace
	if len(h.informerScope) == 0 {
		h.informerScope = metav1.NamespaceAll
	}
	h.informerFactory = informers.NewSharedInformerFactoryWithOptions(
		h.clientset, h.resyncPeriod,
		informers.WithNamespace(h.informerScope),
		informers.WithTweakListOptions(h.tweakListOptions))
}

// SetInformerFactoryTweakListOptions sets a custom filter on all listers of
// the configured SharedInformerFactory.
func (h *Handler) SetInformerFactoryTweakListOptions(tweakListOptions internalinterfaces.TweakListOptionsFunc) {
	h.l.Lock()
	defer h.l.Unlock()
	h.tweakListOptions = tweakListOptions
	if len(h.informerScope) == 0 {
		h.informerScope = metav1.NamespaceAll
	}
	h.informerFactory = informers.NewSharedInformerFactoryWithOptions(
		h.clientset, h.resyncPeriod,
		informers.WithNamespace(h.informerScope),
		informers.WithTweakListOptions(h.tweakListOptions))
}

// InformerFactory returns underlying SharedInformerFactory which provides
// shared informer for resources in all known API group version.
func (h *Handler) InformerFactory() informers.SharedInformerFactory {
	return h.informerFactory
}

// JobInformer returns underlying JobInformer which provides access to a shared
// informer and lister for job.
func (h *Handler) JobInformer() informersbatch.JobInformer {
	return h.informerFactory.Batch().V1().Jobs()
}

// Informer returns underlying SharedIndexInformer which provides add and Indexers
// ability based on SharedInformer.
func (h *Handler) Informer() cache.SharedIndexInformer {
	return h.informerFactory.Batch().V1().Jobs().Informer()
}

// Lister returns underlying JobLister which helps list jobs.
func (h *Handler) Lister() listersbatch.JobLister {
	return h.informerFactory.Batch().V1().Jobs().Lister()
}

// RunInformer start and run the shared informer, returning after it stops.
// The informer will be stopped when stopCh is closed.
//
// AddFunc, updateFunc, and deleteFunc are used to handle add, update,
// and delete event of k8s job resource, respectively.
func (h *Handler) RunInformer(
	stopCh <-chan struct{},
	addFunc func(obj interface{}),
	updateFunc func(oldObj, newObj interface{}),
	deleteFunc func(obj interface{})) {

	h.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	})

	// method 1, recommended
	h.InformerFactory().Start(stopCh)
	logrus.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, h.Informer().HasSynced); !ok {
		logrus.Error("failed to wait for caches to sync")
	}

	//// method 2
	//h.InformerFactory().Start(stopCh)
	//logrus.Info("Waiting for informer caches to sync")
	//h.InformerFactory().WaitForCacheSync(stopCh)

	//// method 3
	//logrus.Info("Waiting for informer caches to sync")
	//h.informerFactory.WaitForCacheSync(stopCh)
	//h.Informer().Run(stopCh)
}

// StartInformer simply call RunInformer.
func (h *Handler) StartInformer(
	stopCh <-chan struct{},
	addFunc func(obj interface{}),
	updateFunc func(oldObj, newObj interface{}),
	deleteFunc func(obj interface{})) {

	h.RunInformer(stopCh, addFunc, updateFunc, deleteFunc)
}
