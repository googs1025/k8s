package persistentvolume

import (
	"encoding/json"
	"errors"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Patch use the default patch type(Strategic Merge Patch) to patch persistentvolume.
// Supported patch types are: "StrategicMergePatchType", "MergePatchType", "JSONPatchType".
//
// For further more Strategic Merge patch, see:
//     https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#before-you-begin
// For a comparison of JSON patch and JSON merge patch, see:
//     https://erosb.github.io/post/json-patch-vs-merge-patch/
func (h *Handler) Patch(original *corev1.PersistentVolume, patch interface{}, patchOptions ...types.PatchType) (*corev1.PersistentVolume, error) {
	switch val := patch.(type) {
	case string:
		var err error
		var patchData []byte
		var jsonData []byte

		if patchData, err = os.ReadFile(val); err != nil {
			return nil, err
		}
		if jsonData, err = yaml.ToJSON(patchData); err != nil {
			return nil, err
		}
		if len(patchOptions) != 0 && patchOptions[0] == types.JSONPatchType {
			return h.jsonPatch(original, jsonData)
		}
		if len(patchOptions) != 0 && patchOptions[0] == types.MergePatchType {
			return h.jsonMergePatch(original, jsonData)
		}
		return h.strategicMergePatch(original, jsonData)

	case []byte:
		var err error
		var jsonData []byte

		if jsonData, err = yaml.ToJSON(val); err != nil {
			return nil, err
		}
		if len(patchOptions) != 0 && patchOptions[0] == types.JSONPatchType {
			return h.jsonPatch(original, jsonData)
		}
		if len(patchOptions) != 0 && patchOptions[0] == types.MergePatchType {
			return h.jsonMergePatch(original, jsonData)
		}
		return h.strategicMergePatch(original, jsonData)

	case *corev1.PersistentVolume:
		return h.diffMergePatch(original, val, patchOptions...)

	case corev1.PersistentVolume:
		return h.diffMergePatch(original, &val, patchOptions...)

	case map[string]interface{}:
		modified := &corev1.PersistentVolume{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(val, modified); err != nil {
			return nil, err
		}
		return h.diffMergePatch(original, modified, patchOptions...)

	case *unstructured.Unstructured:
		modified := &corev1.PersistentVolume{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(val.UnstructuredContent(), modified); err != nil {
			return nil, err
		}
		return h.diffMergePatch(original, modified, patchOptions...)

	case unstructured.Unstructured:
		modified := &corev1.PersistentVolume{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(val.UnstructuredContent(), modified); err != nil {
			return nil, err
		}
		return h.diffMergePatch(original, modified, patchOptions...)

	case runtime.Object:
		modified, ok := patch.(*corev1.PersistentVolume)
		if !ok {
			return nil, errors.New("patch data type is not *corev1.PersistentVolume")
		}
		return h.diffMergePatch(original, modified, patchOptions...)

	default:
		return nil, ErrInvalidPathType
	}
}

// strategicMergePatch use the "Strategic Merge Patch" patch type to patch persistentvolume.
//
// Notice that the patch did not replace the containers list. Instead it added
// a new Container to the list. In other words, the list in the patch was merged
// with the existing list.
//
// This is not always what happens when you use a strategic merge patch on a list.
// In some cases, the list is replaced, not merged.
//
// Note: Strategic merge patch is not supported for custom resources.
// For further more Strategic Merge patch, see:
//     https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#before-you-begin
func (h *Handler) strategicMergePatch(original *corev1.PersistentVolume, patchData []byte) (*corev1.PersistentVolume, error) {
	if len(patchData) == 0 || string(patchData) == "{}" {
		return original, nil
	}
	return h.clientset.CoreV1().PersistentVolumes().
		Patch(h.ctx, original.Name, types.StrategicMergePatchType, patchData, h.Options.PatchOptions)
}

// jsonMergePatch use the "JSON Merge Patch" patch type to patch persistentvolume.
// A JSON merge patch is different from strategic merge patch, With a JSON merge patch,
// If you want to update a list, you have to specify the entire new list.
// And the new list completely replicas the existing list.
//
// For a comparison of JSON patch and JSON merge patch, see:
//     https://erosb.github.io/post/json-patch-vs-merge-patch/
// For further more Json Patch see:
//     https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#before-you-begin
//     https://tools.ietf.org/html/rfc6902
func (h *Handler) jsonMergePatch(original *corev1.PersistentVolume, patchData []byte) (*corev1.PersistentVolume, error) {
	if len(patchData) == 0 || string(patchData) == "{}" {
		return original, nil
	}
	return h.clientset.CoreV1().PersistentVolumes().
		Patch(h.ctx, original.Name, types.StrategicMergePatchType, patchData, h.Options.PatchOptions)
}

// jsonPatch use "JSON Patch" patch type to patch persistentvolume.
//
// For a comparison of JSON patch and JSON merge patch, see:
//     https://erosb.github.io/post/json-patch-vs-merge-patch/
// For further more Json Merge Patch see:
//     https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#before-you-begin
//     https://tools.ietf.org/html/rfc7386
func (h *Handler) jsonPatch(original *corev1.PersistentVolume, patchData []byte) (*corev1.PersistentVolume, error) {
	return h.clientset.CoreV1().PersistentVolumes().Patch(h.ctx,
		original.Name, types.JSONPatchType, patchData, h.Options.PatchOptions)
}

// diffMergePatch will tak the difference data between original and modified persistentvolume object,
// and use the default patch type(Strategic Merge Patch) patch the differen persistentvolume.
// You can set patchOptions to MergePatchType to use the "JSON Merge Patch" to
// patch persistentvolume.
func (h *Handler) diffMergePatch(original, modified *corev1.PersistentVolume, patchOptions ...types.PatchType) (*corev1.PersistentVolume, error) {
	var (
		err          error
		originalJson []byte
		modifiedJson []byte
		patchData    []byte
	)

	if originalJson, err = json.Marshal(original); err != nil {
		return nil, err
	}
	if modifiedJson, err = json.Marshal(modified); err != nil {
		return nil, err
	}
	if patchData, err = strategicpatch.CreateTwoWayMergePatch(originalJson, modifiedJson, corev1.PersistentVolume{}); err != nil {
		return nil, err
	}
	if len(patchData) == 0 || string(patchData) == "{}" {
		return original, nil
	}

	if len(patchOptions) != 0 && patchOptions[0] == types.MergePatchType {
		return h.clientset.CoreV1().PersistentVolumes().
			Patch(h.ctx, original.Name, types.MergePatchType, patchData, h.Options.PatchOptions)
	}
	return h.clientset.CoreV1().PersistentVolumes().
		Patch(h.ctx, original.Name, types.StrategicMergePatchType, patchData, h.Options.PatchOptions)
}
