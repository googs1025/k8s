package main

import (
	"io/ioutil"

	"github.com/forbearing/k8s/deployment"
	"k8s.io/apimachinery/pkg/runtime"
)

func Deployment_Delete() {
	handler, err := deployment.New(ctx, namespace, kubeconfig)
	if err != nil {
		panic(err)
	}
	defer cleanup(handler)

	// 1. delete deployment by name.
	handler.Apply(filename)
	checkErr("delete deployment by name", "", handler.Delete(name))

	// 2. delete deployment from file.
	// You should always explicitly call DeleteFromFile to delete a deployment
	// from file. if the parameter type passed to the `Delete` method is string,
	// the `Delete` method will call DeleteByName not DeleteFromFile.
	handler.Apply(filename)
	checkErr("delete deployment from file", "", handler.DeleteFromFile(filename))

	// 3. delete deployment from bytes.
	var data []byte
	if data, err = ioutil.ReadFile(filename); err != nil {
		panic(err)
	}
	handler.Apply(filename)
	checkErr("delete deployment from bytes", "", handler.Delete(data))

	// 4. delete deployment from *appsv1.deployment
	deploy, _ := handler.Apply(filename)
	checkErr("delete deployment from *appsv1.Deployment", "", handler.Delete(deploy))

	// 5. delete deployment from appsv1.deployment
	deploy, _ = handler.Apply(filename)
	checkErr("delete deployment from appsv1.Deployment", "", handler.Delete(*deploy))

	// 6. delete deployment from runtime.Object.
	deploy, _ = handler.Apply(filename)
	object := runtime.Object(deploy)
	checkErr("delete deployment from runtime.Object", "", handler.Delete(object))

	// 7. delete deployment from unstructured data, aka map[string]interface{}.
	handler.Apply(unstructData)
	checkErr("delete deployment from unstructured data", "", handler.Delete(unstructData))

	// Output:

	//2022/07/22 22:20:18 delete deployment by name success:
	//2022/07/22 22:20:18 delete deployment from file success:
	//2022/07/22 22:20:18 delete deployment from bytes success:
	//2022/07/22 22:20:18 delete deployment from *appsv1.Deployment success:
	//2022/07/22 22:20:18 delete deployment from appsv1.Deployment success:
	//2022/07/22 22:20:18 delete deployment from runtime.Object success:
	//2022/07/22 22:20:19 delete deployment from unstructured data success:
}
