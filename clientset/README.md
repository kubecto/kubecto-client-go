`client-go`支持`4`种`Client`客户端对象与`Kubernetes API Server`交互的方式


                                                                                                                                                                

`RESTClient`是一种最基础的客户端，使用时需要指定`Resource`和`Version`等信息，编写代码时需要提前知道`Resource`所在的`Group`和对应的`Version`信息。相比`RESTClient，ClientSet`使用起来更加便捷，一般情况下，开发者对`Kubernetes`进行二次开发时通常使用`ClientSet`。

`ClientSet`在`RESTClient`的基础上封装了对`Resource`和`Version`的管理方法。每一个`Resource`可以理解为一个客户端，而`ClientSet`则是多个客户端的集合，每一个`Resource`和`Version`都以函数的方式暴露给开发者，例如，`ClientSet`提供的`RbacV1`、`CoreV1`、`NetworkingV1`等接口函数。

注意：`ClientSet`仅能访问`Kubernetes`自身内置的资源（即客户端集合内的资源），不能直接访问`CRD`自定义资源。如果需要`ClientSet`访问`CRD`自定义资源，可以通过`client-gen`代码生成器重新生成`ClientSet`，在`ClientSet`集合中自动生成与`CRD`操作相关的接口。

类似于`kubectl`命令，通过`ClientSet` 创建一个新的命名空间`cto`以及一个新的`deployment`，`ClientSet Example`代码示例如下：

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
    NAMESPACE = "cto"
    DEPLOYMENTNAME = "kubecto-deploy"
    IMAGE = "nginx:1.13"
    PORT = 80
    REPLICAS = 2
)

func main() {
	var kubeconfig *string
        // home是家目录，如果能取得家目录的值，就可以用来做默认值
	if home := homedir.HomeDir(); home != "" {
		// 如果输入了kubeconfig参数，该参数的值就是kubeconfig文件的绝对路径，
		// 如果没有输入kubeconfig参数，就用默认路径~/.kube/config
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) kubeconfig文件的绝对路径")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "kubeconfig文件的绝对路径")
	}
	flag.Parse()

	// 从本机加载kubeconfig配置文件，因此第一个参数为空字符串
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
        // kubeconfig加载失败就直接退出了
	if err != nil {
		panic(err)
	}
        // 实例化clientset对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
		// 引用namespace的函数
		createNamespace(clientset)

		// 引用deployment的函数
		createDeployment(clientset)
}

// 新建namespace
func createNamespace(clientset *kubernetes.Clientset) {
	namespaceClient := clientset.CoreV1().Namespaces()

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: NAMESPACE,
		},
	}

	result, err := namespaceClient.Create(context.TODO(), namespace, metav1.CreateOptions{})

	if err!=nil {
		panic(err.Error())
	}

	fmt.Printf("Create namespace %s \n", result.GetName())
}
// 新建deployment
func createDeployment(clientset *kubernetes.Clientset) {
        //如果希望在default命名空间下场景可以引用apiv1.NamespaceDefault默认字符
	//deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
        //拿到deployment的客户端
	deploymentsClient := clientset.AppsV1().Deployments(NAMESPACE)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: DEPLOYMENTNAME,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(REPLICAS),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "kubecto",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "kubecto",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: IMAGE,
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: PORT,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())
}
//引用replicas带入副本集
func int32Ptr(i int32) *int32 { return &i }
```

运行以上代码，会创建`2`个`nginx`的`deployment`。首先加载`kubeconfig`配置信息，`kubernetes.NewForConfig`通过`kubeconfig`配置信息实例化`clientset`对象，该对象用于管理所有`Resource`的客户端。
