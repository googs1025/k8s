package rolebinding

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Delete deletes rolebinding from type string, []byte, *rbacv1.RoleBinding,
// rbacv1.RoleBinding, runtime.Object, *unstructured.Unstructured,
// unstructured.Unstructured or map[string]interface{}.

// If passed parameter type is string, it will simply call DeleteByName instead of DeleteFromFile.
// You should always explicitly call DeleteFromFile to delete a rolebinding from file path.
func (h *Handler) Delete(obj interface{}) error {
	switch val := obj.(type) {
	case string:
		return h.DeleteByName(val)
	case []byte:
		return h.DeleteFromBytes(val)
	case *rbacv1.RoleBinding:
		return h.DeleteFromObject(val)
	case rbacv1.RoleBinding:
		return h.DeleteFromObject(&val)
	case runtime.Object:
		return h.DeleteFromObject(val)
	case *unstructured.Unstructured:
		return h.DeleteFromUnstructured(val)
	case unstructured.Unstructured:
		return h.DeleteFromUnstructured(&val)
	case map[string]interface{}:
		return h.DeleteFromMap(val)
	default:
		return ERR_TYPE_DELETE
	}
}

// DeleteByName deletes rolebinding by name.
func (h *Handler) DeleteByName(name string) error {
	return h.clientset.RbacV1().RoleBindings(h.namespace).Delete(h.ctx, name, h.Options.DeleteOptions)
}

// DeleteFromFile deletes rolebinding from yaml file.
func (h *Handler) DeleteFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return h.DeleteFromBytes(data)
}

// DeleteFromBytes deletes rolebinding from bytes.
func (h *Handler) DeleteFromBytes(data []byte) error {
	rbJson, err := yaml.ToJSON(data)
	if err != nil {
		return err
	}

	rb := &rbacv1.RoleBinding{}
	err = json.Unmarshal(rbJson, rb)
	if err != nil {
		return err
	}
	return h.deleteRolebinding(rb)
}

// DeleteFromObject deletes rolebinding from runtime.Object.
func (h *Handler) DeleteFromObject(obj runtime.Object) error {
	rb, ok := obj.(*rbacv1.RoleBinding)
	if !ok {
		return fmt.Errorf("object is not *rbacv1.RoleBinding")
	}
	return h.deleteRolebinding(rb)
}

// DeleteFromUnstructured deletes rolebinding from *unstructured.Unstructured.
func (h *Handler) DeleteFromUnstructured(u *unstructured.Unstructured) error {
	rb := &rbacv1.RoleBinding{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), rb)
	if err != nil {
		return err
	}
	return h.deleteRolebinding(rb)
}

// DeleteFromMap deletes rolebinding from map[string]interface{}.
func (h *Handler) DeleteFromMap(u map[string]interface{}) error {
	rb := &rbacv1.RoleBinding{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, rb)
	if err != nil {
		return err
	}
	return h.deleteRolebinding(rb)
}

// deleteRolebinding
func (h *Handler) deleteRolebinding(rb *rbacv1.RoleBinding) error {
	var namespace string
	if len(rb.Namespace) != 0 {
		namespace = rb.Namespace
	} else {
		namespace = h.namespace
	}
	return h.clientset.RbacV1().RoleBindings(namespace).Delete(h.ctx, rb.Name, h.Options.DeleteOptions)
}
