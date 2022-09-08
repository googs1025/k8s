package namespace

import (
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
)

// Watch watch all namespace resources.
//
// Object as the parameter of addFunc, modifyFunc, deleteFunc:
//  * If Event.Type is Added or Modified: the new state of the object.
//  * If Event.Type is Deleted: the state of the object immediately before deletion.
//  * If Event.Type is Bookmark: the object (instance of a type being watched) where
//    only ResourceVersion field is set. On successful restart of watch from a
//    bookmark resourceVersion, client is guaranteed to not get repeat event
//    nor miss any events.
//  * If Event.Type is Error: *api.Status is recommended; other types may make sense
//    depending on context.
func (h *Handler) Watch(addFunc, modifyFunc, deleteFunc func(obj interface{})) error {
	return h.WatchByLabel("", addFunc, deleteFunc, modifyFunc)
}

// WatchByName watch a single namespace reseource.
//
// Object as the parameter of addFunc, modifyFunc, deleteFunc:
//  * If Event.Type is Added or Modified: the new state of the object.
//  * If Event.Type is Deleted: the state of the object immediately before deletion.
//  * If Event.Type is Bookmark: the object (instance of a type being watched) where
//    only ResourceVersion field is set. On successful restart of watch from a
//    bookmark resourceVersion, client is guaranteed to not get repeat event
//    nor miss any events.
//  * If Event.Type is Error: *api.Status is recommended; other types may make sense
//    depending on context.
func (h *Handler) WatchByName(name string, addFunc, modifyFunc, deleteFunc func(obj interface{})) error {
	listOptions := metav1.SingleObject(metav1.ObjectMeta{Name: name})
	listOptions.TimeoutSeconds = new(int64)
	return h.watchNamespace(listOptions, addFunc, modifyFunc, deleteFunc)
}

// WatchByLabel watch a single or multiple Namespace resources selected by the label.
// Multiple labels are separated by ",", label key and value conjunctaed by "=".
//
// Object as the parameter of addFunc, modifyFunc, deleteFunc:
//  * If Event.Type is Added or Modified: the new state of the object.
//  * If Event.Type is Deleted: the state of the object immediately before deletion.
//  * If Event.Type is Bookmark: the object (instance of a type being watched) where
//    only ResourceVersion field is set. On successful restart of watch from a
//    bookmark resourceVersion, client is guaranteed to not get repeat event
//    nor miss any events.
//  * If Event.Type is Error: *api.Status is recommended; other types may make sense
//    depending on context.
func (h *Handler) WatchByLabel(labels string, addFunc, modifyFunc, deleteFunc func(obj interface{})) error {
	return h.watchNamespace(metav1.ListOptions{LabelSelector: labels, TimeoutSeconds: new(int64)},
		addFunc, modifyFunc, deleteFunc)
}

// WatchByField watch a single or multiple Namespace resources selected by the field.
//
// Object as the parameter of addFunc, modifyFunc, deleteFunc:
//  * If Event.Type is Added or Modified: the new state of the object.
//  * If Event.Type is Deleted: the state of the object immediately before deletion.
//  * If Event.Type is Bookmark: the object (instance of a type being watched) where
//    only ResourceVersion field is set. On successful restart of watch from a
//    bookmark resourceVersion, client is guaranteed to not get repeat event
//    nor miss any events.
//  * If Event.Type is Error: *api.Status is recommended; other types may make sense
//    depending on context.
func (h *Handler) WatchByField(field string, addFunc, modifyFunc, deleteFunc func(obj interface{})) error {
	fieldSelector, err := fields.ParseSelector(field)
	if err != nil {
		return err
	}
	listOptions := metav1.ListOptions{FieldSelector: fieldSelector.String(), TimeoutSeconds: new(int64)}
	return h.watchNamespace(listOptions, addFunc, modifyFunc, deleteFunc)
}

// watchNamespace watch namespace resources according to listOptions.
func (h *Handler) watchNamespace(listOptions metav1.ListOptions,
	addFunc, modifyFunc, deleteFunc func(obj interface{})) (err error) {

	var watcher watch.Interface
	// if event channel is closed, it means the server has closed the connection,
	// reconnect to kubernetes API server.
	for {
		if watcher, err = h.clientset.CoreV1().Namespaces().Watch(h.ctx, listOptions); err != nil {
			return err
		}
		// kubernetes retains the resource event history, which includes this
		// initial event, so that when our program first start, we are automatically
		// notified of the namespace existence and current state.
		// There we will not ignore the first resource added event.
		for event := range watcher.ResultChan() {
			switch event.Type {
			case watch.Added:
				addFunc(event.Object)
			case watch.Modified:
				modifyFunc(event.Object)
			case watch.Deleted:
				deleteFunc(event.Object)
			case watch.Bookmark:
				log.Debug("watch namespace: bookmark")
			case watch.Error:
				log.Debug("watch namespace: error")
			}
		}
		// If event channel is closed, it means the server has closed the connection
		log.Debug("watch namespace: reconnect to kubernetes")
		watcher.Stop()
	}
}
