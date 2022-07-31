package service

import (
	"time"

	"k8s.io/client-go/informers"
	informerscore "k8s.io/client-go/informers/core/v1"
	listerscore "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// SetInformerResyncPeriod will set informer resync period.
func (h *Handler) SetInformerResyncPeriod(resyncPeriod time.Duration) {
	h.informerFactory = informers.NewSharedInformerFactory(h.clientset, resyncPeriod)
}

// InformerFactory returns underlying SharedInformerFactory which provides
// shared informer for resources in all known API group version.
func (h *Handler) InformerFactory() informers.SharedInformerFactory {
	return h.informerFactory
}

// ServiceInformer returns underlying ServiceInformer which provides access to
// a shared informer and lister for service.
func (h *Handler) ServiceInformer() informerscore.ServiceInformer {
	return h.informerFactory.Core().V1().Services()
}

// Informer returns underlying SharedIndexInformer which provides add and Indexers
// ability based on SharedInformer.
func (h *Handler) Informer() cache.SharedIndexInformer {
	return h.informerFactory.Core().V1().Services().Informer()
}

// Lister returns underlying ServiceLister which helps list services.
func (h *Handler) Lister() listerscore.ServiceLister {
	return h.informerFactory.Core().V1().Services().Lister()
}

// RunInformer
func (h *Handler) RunInformer(
	addFunc func(obj interface{}),
	updateFunc func(oldObj, newObj interface{}),
	deleteFunc func(obj interface{}),
	stopCh chan struct{}) {
	h.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    addFunc,
		UpdateFunc: updateFunc,
		DeleteFunc: deleteFunc,
	})
	h.Informer().Run(stopCh)
}
