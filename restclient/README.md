`DynamicClient`是一种动态客户端，它可以对任意`Kubernetes`资源进行`RESTful`操作，包括`CRD`自定义资源。`DynamicClient`与`ClientSet`操作类似，同样封装了`RESTClient`，同样提供了`Create、Update、Delete、Get、List、Watch、Patch`等方法。

`DynamicClient`与`ClientSet`最大的不同之处是，`ClientSet`仅能访问`Kubernetes`自带的资源（即客户端集合内的资源），不能直接访问`CRD`自定义资源。`ClientSet`需要预先实现每种`Resource`和`Version`的操作，其内部的数据都是结构化数据（即已知数据结构）。而`DynamicClient`内部实现了`Unstructured`，用于处理非结构化数据结构（即无法提前预知数据结构），这也是`DynamicClient`能够处理`CRD`自定义资源的关键。

注意：`DynamicClient`不是类型安全的，因此在访问`CRD`自定义资源时需要特别注意。例如，在操作指针不当的情况下可能会导致程序崩溃。

`DynamicClient`的处理过程将`Resource`（例如`PodList`）转换成`Unstructured`结构类型，`Kubernetes`的所有`Resource`都可以转换为该结构类型。处理完成后，再将`Unstructured`转换成`PodList`。整个过程类似于`Go`语言的`interface{}`断言转换过程。另外，`Unstructured`结构类型是通过`map[string]interface{}`转换的。

类似于`kubectl`命令，通过`DynamicClient` 创建`deployment`,并使用`list`列出当前`pod`的名称和数量，`DynamicClient Example`代码示例如下：

```go
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
// home是家目录，如果能取得家目录的值，就可以用来做默认值
		// 如果输入了kubeconfig参数，该参数的值就是kubeconfig文件的绝对路径，
		// 如果没有输入kubeconfig参数，就用默认路径~/.kube/config
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

//定义函数内的变量
        namespace := "default"
        replicas := 2
        deployname := "ku"
        image := "nginx:1.17"

	// 从本机加载kubeconfig配置文件，因此第一个参数为空字符串
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
        // kubeconfig加载失败就直接退出了
	if err != nil {
		panic(err)
	}
        // dynamic.NewForConfig实例化对象
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
        //使用schema的包带入gvr
	deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
        //定义结构化数据结构
	deployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name": deployname,
			},
			"spec": map[string]interface{}{
				"replicas": replicas,
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "demo",
					},
				},
				"template": map[string]interface{}{
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							"app": "demo",
						},
					},

					"spec": map[string]interface{}{
						"containers": []map[string]interface{}{
							{
								"name":  "web",
								"image": image,
								"ports": []map[string]interface{}{
									{
										"name":          "http",
										"protocol":      "TCP",
										"containerPort": 80,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 创建 Deployment
	fmt.Println("创建 deployment...")
	result, err := client.Resource(deploymentRes).Namespace(namespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("创建 deployment %q.\n", result.GetName())

	// 列出 Deployments
	prompt()
	fmt.Printf("在命名空间中列出deployment %q:\n", apiv1.NamespaceDefault)
	list, err := client.Resource(deploymentRes).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		replicas, found, err := unstructured.NestedInt64(d.Object, "spec", "replicas")
		if err != nil || !found {
			fmt.Printf("Replicas not found for deployment %s: error=%s", d.GetName(), err)
			continue
		}
		fmt.Printf(" * %s (%d replicas)\n", d.GetName(), replicas)
	}

}
func prompt() {
	fmt.Printf("--------------> 按回车键继续 <--------------.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
```

首先加载`kubeconfig`配置信息，`dynamic.NewForConfig`通过`kubeconfig`配置信息实例化`dynamicClient`对象，该对象用于管理`Kubernetes`的所有`Resource`的客户端，例如对`Resource`执行`Create、Update、Delete、Get、List、Watch、Patch`等操作。
