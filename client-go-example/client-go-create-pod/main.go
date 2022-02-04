package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
 //读取kubeconfig文件
  rules := clientcmd.NewDefaultClientConfigLoadingRules()
  kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
  config, err := kubeconfig.ClientConfig()
  if err != nil {
    panic(err)
  }
  
  //实例化clientst
  clientset := kubernetes.NewForConfigOrDie(config)

  //pod模版
  newPod := &corev1.Pod{
    ObjectMeta: metav1.ObjectMeta{
      Name: "test-pod",
    },
    Spec: corev1.PodSpec{
      Containers: []corev1.Container{
        {Name: "busybox", Image: "busybox:latest", Command: []string{"sleep", "100000"}},
      },
    },
  }

 //创建pod
  pod, err := clientset.CoreV1().Pods("default").Create(context.Background(), newPod, metav1.CreateOptions{})
  if err != nil {
    panic(err)
  }
    fmt.Printf("Created pod %q.\n", pod.GetObjectMeta().GetName())
}
