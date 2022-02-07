package main

import (
	"flag"
	"fmt"
	"time"

	"k8s.io/klog/v2"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// Controller演示了如何用client-go实现一个控制器。
type Controller struct {
	indexer  cache.Indexer
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

// NewController创建一个新的Controller。
func NewController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *Controller {
	return &Controller{
		informer: informer,
		indexer:  indexer,
		queue:    queue,
	}
}

func (c *Controller) processNextItem() bool {
	// 等待，直到工作队列中有一个新项
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	//告诉队列我们已经处理完这个键。这允许安全的并行处理，因为两个具有相同密钥的pod永远不会并行处理。
	defer c.queue.Done(key)

	// 调用包含业务逻辑的方法
	err := c.syncToStdout(key.(string))
	// 如果在执行业务逻辑期间出现错误，则处理错误
	c.handleErr(err, key)
	return true
}

// syncToStdout是控制器的业务逻辑。在这个控制器中，它只是将关于pod的信息打印到stdout。
//在发生错误的情况下，它必须简单地返回错误。
//重试逻辑不应该是业务逻辑的一部分。
func (c *Controller) syncToStdout(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		// 下面我们将用一个Pod来暖化我们的缓存，这样我们将看到一个Pod的删除
		fmt.Printf("Pod %s does not exist anymore\n", key)
	} else {
		//注意，如果你有一个本地控制资源，你也必须检查uid，这是依赖于实际的实例，以检测一个Pod被重新创建了相同的名称
		fmt.Printf("Sync/Add/Update for Pod %s\n", obj.(*v1.Pod).GetName())
	}
	return nil
}

// handleErr检查是否发生了错误，并确保稍后重试。
func (c *Controller) handleErr(err error, key interface{}) {
	if err == nil {
		//忘记#AddRateLimited历史的键在每次成功的同步。
		//这确保了以后对这个键的更新的处理不会因为过时的错误历史而延迟。
		c.queue.Forget(key)
		return
	}

	// 如果出现问题，这个控制器会重试5次。在那之后，它就会停止尝试。
	if c.queue.NumRequeues(key) < 5 {
		klog.Infof("Error syncing pod %v: %v", key, err)

		//重新排队键速率限制。
		//根据队列上的速率限制器和重新排队的历史记录，稍后将再次处理该键。
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)
	// 向外部实体报告，即使在多次重试之后，我们仍不能成功处理此密钥
	runtime.HandleError(err)
	klog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

// Run开始观察和同步。
func (c *Controller) Run(workers int, stopCh chan struct{}) {
	defer runtime.HandleCrash()

	// 我们完工后让任务停下来
	defer c.queue.ShutDown()
	klog.Info("Starting Pod controller")

	go c.informer.Run(stopCh)

	// 在开始处理队列中的项目之前，等待所有涉及的缓存被同步
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Info("Stopping Pod controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func main() {
	var kubeconfig string
	var master string

	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.Parse()

	// 创建连接
	config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
	if err != nil {
		klog.Fatal(err)
	}

	// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatal(err)
	}

	// 创建 pod watcher
	podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())

	// 创建 workqueue
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	//在informer的帮助下绑定工作队列到缓存通过这种方式，我们可以确保每当缓存被更新时，pod键都被添加到工作队列中。
	//请注意，当我们最终处理工作队列中的项目时，我们可能会看到一个比负责触发更新的版本更新的Pod。
	indexer, informer := cache.NewIndexerInformer(podListWatcher, &v1.Pod{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// IndexerInformer使用一个增量队列，因此我们必须使用这个键函数进行删除。
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				queue.Add(key)
			}
		},
	}, cache.Indexers{})

	controller := NewController(queue, indexer, informer)

	//我们现在可以为初始同步预热缓存。
	//让我们假设我们在上次运行时知道一个pod“mypod”，因此将它添加到缓存中。
	//如果这个pod不再存在，控制器将在缓存同步后收到移除的通知。

	indexer.Add(&v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "test-node-local-dns",
			Namespace: v1.NamespaceDefault,
		},
	})

	//现在可以启动这个controller
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// 永远等待
	select {}
}
