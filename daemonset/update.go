package daemonset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Update updates daemonset from type string, []byte, *appsv1.DaemonSet,
// appsv1.DaemonSet, runtime.Object or map[string]interface{}.
func (h *Handler) Update(obj interface{}) (*appsv1.DaemonSet, error) {
	switch val := obj.(type) {
	case string:
		return h.UpdateFromFile(val)
	case []byte:
		return h.UpdateFromBytes(val)
	case *appsv1.DaemonSet:
		return h.UpdateFromObject(val)
	case appsv1.DaemonSet:
		return h.UpdateFromObject(&val)
	case runtime.Object:
		return h.UpdateFromObject(val)
	case map[string]interface{}:
		return h.UpdateFromUnstructured(val)
	default:
		return nil, ERR_TYPE_UPDATE
	}
}

// UpdateFromFile updates daemonset from yaml file.
func (h *Handler) UpdateFromFile(filename string) (*appsv1.DaemonSet, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return h.UpdateFromBytes(data)
}

// UpdateFromBytes updates daemonset from bytes.
func (h *Handler) UpdateFromBytes(data []byte) (*appsv1.DaemonSet, error) {
	dsJson, err := yaml.ToJSON(data)
	if err != nil {
		return nil, err
	}

	ds := &appsv1.DaemonSet{}
	err = json.Unmarshal(dsJson, ds)
	if err != nil {
		return nil, err
	}
	return h.updateDaemonset(ds)
}

// UpdateFromObject updates daemonset from runtime.Object.
func (h *Handler) UpdateFromObject(obj runtime.Object) (*appsv1.DaemonSet, error) {
	ds, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return nil, fmt.Errorf("object is not *appsv1.DaemonSet")
	}
	return h.updateDaemonset(ds)
}

// UpdateFromUnstructured updates daemonset from map[string]interface{}.
func (h *Handler) UpdateFromUnstructured(u map[string]interface{}) (*appsv1.DaemonSet, error) {
	ds := &appsv1.DaemonSet{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, ds)
	if err != nil {
		return nil, err
	}
	return h.updateDaemonset(ds)
}

// updateDaemonset
func (h *Handler) updateDaemonset(ds *appsv1.DaemonSet) (*appsv1.DaemonSet, error) {
	var namespace string
	if len(ds.Namespace) != 0 {
		namespace = ds.Namespace
	} else {
		namespace = h.namespace
	}
	ds.ResourceVersion = ""
	ds.UID = ""
	return h.clientset.AppsV1().DaemonSets(namespace).Update(h.ctx, ds, h.Options.UpdateOptions)
}
