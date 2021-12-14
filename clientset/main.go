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
