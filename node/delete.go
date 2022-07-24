package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Delete deletes node from type string, []byte, *corev1.Node,
// corev1.Node, runtime.Object or map[string]interface{}.

// If passed parameter type is string, it will simply call DeleteByName instead of DeleteFromFile.
// You should always explicitly call DeleteFromFile to delete a node from file path.
func (h *Handler) Delete(obj interface{}) error {
	switch val := obj.(type) {
	case string:
		return h.DeleteByName(val)
	case []byte:
		return h.DeleteFromBytes(val)
	case *corev1.Node:
		return h.DeleteFromObject(val)
	case corev1.Node:
		return h.DeleteFromObject(&val)
	case runtime.Object:
		return h.DeleteFromObject(val)
	case map[string]interface{}:
		return h.DeleteFromUnstructured(val)
	default:
		return ERR_TYPE_DELETE
	}
}

// DeleteByName deletes node by name.
func (h *Handler) DeleteByName(name string) error {
	return h.clientset.CoreV1().Nodes().Delete(h.ctx, name, h.Options.DeleteOptions)
}

// DeleteFromFile deletes node from yaml file.
func (h *Handler) DeleteFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return h.DeleteFromBytes(data)
}

// DeleteFromBytes deletes node from bytes.
func (h *Handler) DeleteFromBytes(data []byte) error {
	nodeJson, err := yaml.ToJSON(data)
	if err != nil {
		return err
	}

	node := &corev1.Node{}
	err = json.Unmarshal(nodeJson, node)
	if err != nil {
		return err
	}
	return h.deleteNode(node)
}

// DeleteFromObject deletes node from runtime.Object.
func (h *Handler) DeleteFromObject(obj runtime.Object) error {
	node, ok := obj.(*corev1.Node)
	if !ok {
		return fmt.Errorf("object is not *corev1.Node")
	}
	return h.deleteNode(node)
}

// DeleteFromUnstructured deletes node from map[string]interface{}.
func (h *Handler) DeleteFromUnstructured(u map[string]interface{}) error {
	node := &corev1.Node{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, node)
	if err != nil {
		return err
	}
	return h.deleteNode(node)
}

// deleteNode
func (h *Handler) deleteNode(node *corev1.Node) error {
	return h.clientset.CoreV1().Nodes().Delete(h.ctx, node.Name, h.Options.DeleteOptions)
}
