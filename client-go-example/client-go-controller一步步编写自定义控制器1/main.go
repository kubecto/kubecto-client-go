package main

import (
	"flag"
	"fmt"
        "path/filepath"
        "time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
        "k8s.io/client-go/util/homedir"
        "k8s.io/client-go/informers"

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

//我们将调用c.run，与channel进行调用,创建channel并启动informer
	ch := make(chan struct{})
	informers := informers.NewSharedInformerFactory(clientset, 10*time.Minute)

//目前我们已经有很好的骨架，我们需要继续前进，在`main.go`
//定义新的控制器，它期望客户端设置并期望部署前，所以需要部署`deploy`部署公式，像这样
	c := newController(clientset, informers.Apps().V1().Deployments())

	informers.Start(ch)
	c.run(ch)
	fmt.Println(informers)
}
