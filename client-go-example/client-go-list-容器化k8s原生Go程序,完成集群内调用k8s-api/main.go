package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// 创建 in-cluster config使用集群内身份验证
        //这里的rest.InClusterConfig实际会调用client-go的服务账户信息,每创建pod的时候都会使用此认证,简单说就是拿服务令牌去连接集群
        //https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/client-go/rest/config.go#L512
        //	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	//	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
        //实例化clientset对象
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		// handle error
		fmt.Printf("error %s, creating clientset\n", err.Error())
	}
	ctx := context.Background()
        fmt.Println("获取default namespace下的pod")

        //1、获取default下pod的名字
	pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
	if err != nil {
		// handle error
		fmt.Printf("error %s, while listing all the pods from default namespace\n", err.Error())
	}

        //在k8s源码当中https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/core/v1/types.go
        //type PodList struct 定义了podlist ,Items []Pod定义了包含切片Pod的项目
	//metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	//metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of pods.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	//Items []Pod `json:"items" protobuf:"bytes,2,rep,name=items"`
       
        //所以我们就可以之间使用for循环取出此切片的长度了
	for _, pod := range pods.Items {
		fmt.Printf("%s\n", pod.Name)
	}

        
	fmt.Println("获取default namespace下的deployment的名字 ")
        //2、获取default下deployment的资源名字
	deployments, err := clientset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})
	if err != nil {
                fmt.Printf("listing deployments %s\n", err.Error())
	}
	for _, d := range deployments.Items {
		fmt.Printf("%s\n", d.Name)
	}

        fmt.Println("获取kube-system namespace下的daemonset的名字 ")       
        //3、获取kube-system下daemonset的资源名字
        daemonsets, err := clientset.AppsV1().DaemonSets("kube-system").List(ctx, metav1.ListOptions{})
        if err != nil {
                fmt.Printf("listing daemonsets %s\n", err.Error())
        }
        for _, ds := range daemonsets.Items {
                fmt.Printf("%s\n", ds.Name)
        }


        fmt.Println("获取kube-system svc")
        //4、获取kube-system下的service
        svc, err := clientset.CoreV1().Services("kube-system").List(ctx, metav1.ListOptions{})
        if err != nil {
                fmt.Printf("listing service %s\n", err.Error())
        }
        for _, service := range svc.Items {
                fmt.Printf("%s\n", service.Name)
        }

}
