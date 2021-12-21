package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
        "k8s.io/apimachinery/pkg/runtime/schema"
        apiv1 "k8s.io/api/core/v1"
        "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)


func main() {
        // home是家目录，如果能取得家目录的值，就可以用来做默认值
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		// 如果输入了kubeconfig参数，该参数的值就是kubeconfig文件的绝对路径，
		// 如果没有输入kubeconfig参数，就用默认路径~/.kube/config
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
//定义函数内的变量名称
        deletedeployname := "ku"
        namespace := "default"
//从本机加载kubeconfig配置文件，因此第一个参数为空字符串
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
//实例化clientset对象
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
//使用schema包带入gvr
        deploymentRes := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

	// 删除 Deployment
	prompt()
	fmt.Println("删除 deployment 中...")
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	if err := client.Resource(deploymentRes).Namespace(namespace).Delete(context.TODO(), deletedeployname, deleteOptions); err != nil {
		panic(err)
	}

	fmt.Println("成功完成删除 deployment，请等待terminating之后再执行回车，列出当前的kubectl get po -n 命名空间.")


	// 查看 Deployments
	prompt()
	fmt.Printf("列出当前命名空间的Pod %q:\n", apiv1.NamespaceDefault)
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

	fmt.Println("------------>按回车键继续")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
