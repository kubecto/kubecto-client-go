package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)
//如果希望在这里初始化这些字段，那么可以使用clientset轻松的从main中获取客户端集
//使用depLister从informer列出部署
//从informer本身同步部署缓存和队列
//使用workqueue的接口来初始化队列
type controller struct {
	clientset      kubernetes.Interface
	depLister      appslisters.DeploymentLister
	depCacheSynced cache.InformerSynced
	queue          workqueue.RateLimitingInterface
}

func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
	c := &controller{
		clientset:      clientset,
		depLister:      depInformer.Lister(),
		depCacheSynced: depInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kubecto-expose"),
	}

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handleAdd,
			DeleteFunc: handleDel,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("starting controller")
	if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
		fmt.Print("waiting for cache to be synced\n")
	}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch
}

func (c * controller) worker () {

}

func handleAdd(obj interface{}) {
	fmt.Println("add was called")

}

func handleDel(obj interface{}) {
	fmt.Println("del was called")

}
