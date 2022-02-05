package main

import (
	"context"
	"fmt"
        "path/filepath"
        "flag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
        "k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 使用kubeconfig中的当前上下文,加载配置文件
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

  //pod模版
  newPod := &corev1.Pod{
    ObjectMeta: metav1.ObjectMeta{
      Name: "test-busybox",
    },
    Spec: corev1.PodSpec{
      Containers: []corev1.Container{
        {Name: "busybox", Image: "busybox:latest", Command: []string{"sleep", "1000"}},
      },
    },
  }

 //创建pod
  pod, err := clientset.CoreV1().Pods("kube-system").Create(context.Background(), newPod, metav1.CreateOptions{})
  if err != nil {
    panic(err)
  }
    fmt.Printf("Created pod %q.\n", pod.GetObjectMeta().GetName())
}
