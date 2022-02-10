package main

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

//一旦我们准备好一个控制器结构，让我们继续创建一个函数，在调用时返回一个控制器，以便我们可以从main调用该函数
func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
	c := &controller{
		clientset:      clientset,                        //使用clientset来进行初始化
		depLister:      depInformer.Lister(),             //将是部署列出
		depCacheSynced: depInformer.Informer().HasSynced, //注册缓存同步信息
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "kubecto-expose"),
		//queue初始化工作队列，设置限速队列，默认限速队列，另外传递的名称为kubectoexpose
	}

	//一旦我们有了部署通知器，我们将添加前面图中讨论的注册函数
	//这里定义了一个可以调用`handleAdd`以及`handleDel`的函数

	//目前我们增加了两个函数一个是添加函数另外一个是删除函数，我们现在将能够调用这两个函数
	//当然这两个函数将会在`deployment`添加或者删除时将会调用
	//如果我们继续实现此功能，首先我们等待`informer`缓存，因为`informer`维护本地缓存，
	//因为我们需要确保信息已成功同步才行，所以我们需要做的是缓存定等待缓存
	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDel,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("starting controller")
	if !cache.WaitForCacheSync(ch, c.depCacheSynced) { //需要传递给连接的informer
		//另外做了个if判断，如果没有这样做，有问题在等待缓存同步时将出现了问题
		fmt.Print("waiting for cache to be synced\n")
	}

	go wait.Until(c.worker, 1*time.Second, ch)
	//等待直到它的作用是它在每个持续时间后调用一个特定函数，直到该特定channel关闭，所以如果
	//这个函数运行我们要指定的函数，这个函数将在每个周期之后运行，直到关闭
	//所以如果我们不关闭这个channel,我们将通过这个函数每次我们指定之后都会贝调用，所以我们
	//称它为调用函数c.worker,指定时间秒，这样我们指定的通道就会将在这里运行，这就是一个很好的例程
	<-ch //定义始终运行，并且监听所有资源
}

func (c *controller) worker() {
	for c.proccessItem() {

	}

}

func (c *controller) proccessItem() bool {
	//无限调用这个函数 此函数将从队列中获取值
	item, shutdown := c.queue.Get() //在队列中获取项目，并关闭
	if shutdown {
		return false //在这种情况关闭是真的，所以我们必须返回false
		//如果是shutdown是不正确的，我们会将这个对象放入这个函数当中
	}

	defer c.queue.Forget(item)
	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("getting key from cache %s\n", err.Error())
	}
	//引入名称和命名空间
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("sllitting key into namespace and name %s\n", err.Error())
		return false
	}

	//现在将为这个特定的控制器来创建service,那么第一步首先来调用deployment

	err = c.syncDeployment(ns, name)
	if err != nil {
		fmt.Printf("syncing deployment %s\n", err.Error())
		return false
	}
	return true //如果按预期运行就返回true
}

func (c *controller) syncDeployment(ns, name string) error {
	ctx := context.Background() //创建上下文，并引入模版当中
	//让我们得到deployment
	//从命名空间当中列出deplpyment和name
	dep, err := c.depLister.Deployments(ns).Get(name)
	if err != nil {
		fmt.Printf("getting deployment from listering %s\n", err.Error())
	}

	//创建service
	//https://pkg.go.dev/k8s.io/api/core/v1,找到对应的包导入,编写svc的模版
	//定义service需要指定服务的名称，以及在哪个命名空间中创建

	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dep.Name,
			Namespace: ns,
		},
		//另外还需要指定端口，以spec当中的规范进行引导定义指定端口
		Spec: corev1.ServiceSpec{
			Selector: depLabels(*dep),
                        Type: corev1.ServiceTypeNodePort,//如果希望是clusterip可以去掉此行
			Ports: []corev1.ServicePort{
				{
					Name: "http",
					Port: 80,
                                        NodePort: 30090,//如果希望是clusterip可以去掉此行
				},
			},
		},
	}
	_, err = c.clientset.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{}) //引入metav1的包
	if err != nil {                                                                      //检查错误
		fmt.Printf("create service %s\n", err.Error())
	}
	return nil

}

func depLabels(dep appsv1.Deployment) map[string]string {
	return dep.Spec.Template.Labels

}

func (c *controller) handleAdd(obj interface{}) {
	//使这些函数成为控制器的方法，显然这样我们就可以访问控制器
	fmt.Println("注册的添加函数已经被调用，创建deployment,则使用此接口输出")
	c.queue.Add(obj) //在队列当中添加此函数

}

func (c *controller) handleDel(obj interface{}) { //
	//使这些函数成为控制器的方法，显然这样我们就可以访问控制器
	fmt.Println("注册的删除函数已经被调用，删除deployment,则使用此接口输出")
	c.queue.Add(obj) //在队列当中添加此函数

}

