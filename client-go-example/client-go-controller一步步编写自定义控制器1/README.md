本次视频我们将研究如何编写`kubernetes`自定义控制器

自定义控制器，这个特定视频的先决条件，你需要在`k8s`开发进阶二学习完`client-go`与`kubernetes  apiserver`通信的应用程序的视频，才能更好的去吸收此次的内容

现在我们将要解决的用例或者说我们将要去编写的自定义控制器

让我们去假设我们现在有一个`kubernetes`集群

![未命名文件(1)](https://user-images.githubusercontent.com/94602819/153161710-2cc1509d-0fa2-46c5-8121-04fc5b208bb8.png)


在这个`k8s`集群部署一个这样的应用程序，假设我们说有一个`base-1`的`java`应用，用户希望能够访问这个特定的应用程序，我们必须做的我们必须要创建一个`service`和一个`ingress`，以便这个用户能够访问部署在`k8s`集群上的应用，如果你不知道这个特定的过程，可以查看一下官方的文档，这些相信大部分的运维人员都会使用它

正如我们现在讨论的有两种方法，两个步骤一个是创建`svc`,一个是创建`ingress`,在这两个步骤之后，便能够访问到我们的应用程序

而现在我们的自定义控制器要做的是我们要访问的控制器，创建部署后立即写入我们的控制器当中，此控制器自动创建相应的`svc`以及`ingress`，而你需要做的就是在`k8s`集群上部署一个应用程序，应用程序将自动暴露给外界，这就是我们此视频自定义控制器要做的事情


![未命名文件(2)](https://user-images.githubusercontent.com/94602819/153161799-363f21a6-f312-4111-a8be-c929d5275ff2.png)


我们称之为cto-exposed，因为它会自动去创建对应的svc以及ingress用于暴露我们的应用，现在我们知道了这个示例的想法了，简而言之，这种方式可以帮助我们自定义控制器的方式去创建并暴露我们的应用，而不是使用传统方式声明一个yaml文件去创建暴露我们的应用


![未命名文件(3)](https://user-images.githubusercontent.com/94602819/153161869-62628551-0c09-4832-881a-ad33188190b2.png)


现在我们开始去尝试研究这个控制器的构建块，我们应该以某种方式或者正在使用的控制器的代码编写，应该以某种方式知道何时创建了特定部署资源，然后才应该创建或者应该公开该部署的资源，所以我们应该可以使用watch,并告知给informer,一旦我们有的了informer,我们将要做的是，我们将注册一些功能，以便deployment在集群上添加或者更新部署后，这些注册的功能是将被调用，所以我们将注册一个函数，假设我们有一个hardadd添加类似的和hardupdate和harddelete,一旦部署在k8s集群上添加更新或者删除，这些函数就会被调用，已经被调用的，
他们要做的就是在数据结构上创建一个队列，我们将作为控制器的一部分，维护该队列


![未命名文件(4)](https://user-images.githubusercontent.com/94602819/153162006-f5465814-6bc2-4856-aef9-fef1db534c8d.png)

所以假设这个函数显然是用对象`obj`调用的，如果你说过的是关于处理部署的添加，添加的部署，将被传递给这个特定的处理程序函数，将把它添加到这个数据结构中

我们现在维护的对列中还有另外一个函数或另一个例程，我们会有这个例程，将从我们正在维护的这个队列中获取对象，然后实际上将在这里执行所有的业务逻辑，所以说，这就是我们要做的这个东西，所以在这个逻辑当中我们将创建`service`和`ingress`,以便添加的部署可以自动的公开，所以这些都是我们将用来编写此内容的构建块，以及我们正在谈论的自定义控制器


![未命名文件(5)](https://user-images.githubusercontent.com/94602819/153162066-95d9d9dd-b3ba-4b48-b7f2-6e817720b227.png)


### 开始编写项目


```
go mod init cto-exposed

[root@kubecto ctoexposed]# cat main.go
package main

import (
	"flag"
	"fmt"
        "path/filepath"

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
        fmt.Println(clientset)
}


[root@kubecto ctoexposed]# go run main.go
&{0xc0003f1ca0 0xc0003d4990 0xc0003d49f0 0xc0003d4a50 0xc0003d4ab0 0xc0003d4b10 0xc0003d4b70 0xc0003d4bd0 0xc0003d4c30 0xc0003d4c90 0xc0003d4cf0 0xc0003d4d50 0xc0003d4db0 0xc0003d4e10 0xc0003d4e70 0xc0003d4ed0 0xc0003d4f30 0xc0003d4f90 0xc0003d4ff0 0xc0003d5050 0xc0003d50b0 0xc0003d5110 0xc0003d5170 0xc0003d51d0 0xc0003d5230 0xc0003d5290 0xc0003d52f0 0xc0003d5350 0xc0003d53b0 0xc0003d5410 0xc0003d5470 0xc0003d54d0 0xc0003d5530 0xc0003d5590 0xc0003d55f0 0xc0003d5650 0xc0003d56b0 0xc0003d5710 0xc0003d5770
```



目前我们已经调试好了我们需要配置的建立与`k8s API`的通信连接，现在我们必须创建用于部署的informer以及我们将要做的事情就是创建shared informer factory 

对于`shared informer factory` 我们可以尝试在代码中这样做，以便`informer`启动，

```
// 创建clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	informers, err := informers.NewSharedInformerFactory(clientset, 10*time.Minute)
        if err != nil {
            fmt.Print("getting informer factory %s\n", err,Error())
        }

	fmt.Println(clientset)
}
```

新建`controller.go`

为了简单起见，让我们看看，我将在同一个包中创建所有这些源文件，但是在实际的生产控制器中你会看到`controller`在单独的包中，所以让我们开始创建`controller`

这就是我们将要执行的控制器将拥有我们讨论过的所有方法以及我们讨论过的例程.


如果我们讨论这个控制器将要拥有的字段，那么首先我们需要设置客户端与`kubernetes`集群进行交互，例如我们一旦拥有这个obj这个对象，我们将创建该对象，或者我们在`k8s`集群中创建资源，来完成我们所做的事情，需要的是我们需要客户端一旦我们设置来客户端

它将是`kubernetes interface`,另外我们还需要`lister,lister`是`informer`的组件，我们使用它再次获取资源。
一旦我们列出来部署，我们就有办法确定是否通知缓存，维护一个缓存，如果该缓存已同步或该缓存已更新，因此我们将其称为部署缓存已同步，认为类型是非正式的`start has`,另外我们还需要一个队列，正如我们上述的图中所描述的那样，一旦调用来这些注册函数，我们就会添加该对象

```
package main

import (
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/tools/cache"
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
```

所以一旦我们准备好一个控制器结构，让我们继续创建一个函数，在调用时返回一个控制器，以便我们可以从`main`

调用该函数

```
func newController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
//所以我们尝试这么做，我们希望clientset kubernetes.Interface，以及deinformer部署informer的应用程序depInformer appsinformers.DeploymentInformer
//
	c := &controller{
		clientset:      clientset,//使用clientset来进行初始化
		depLister:      depInformer.Lister(),//将是部署列出
		depCacheSynced: depInformer.Informer().HasSynced,//注册缓存同步将是部署信息
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ctoexpose"),
     //queue初始化工作队列，设置限速队列，默认限速队列，另外传递的名称为ctoexpose
	}
```

一旦我们有了部署通知器，我们将添加前面图中讨论的注册函数

这里定义了一个可以调用`handleAdd`以及`handleDel`的函数

```
depInformer.Informer().AddEventHandler(
                cache.ResourceEventHandlerFuncs{
                        AddFunc:    handleAdd,
                        DeleteFunc: handleDel,
                },
        )
```

接着定义类型为`handleAdd/handleDel`接口的对象。

目前我们增加了两个函数一个是添加函数另外一个是删除函数，我们现在将能够调用这两个函数

当然这两个函数将会在`deployment`添加或者删除时将会调用

如果我们继续实现此功能，首先我们等待`informer`缓存，因为`informer`维护本地缓存，

因为我们需要确保信息已成功同步才行，所以我们需要做的是缓存定等待缓存


```
func (c *controller) run(ch <-chan struct{}) {
        fmt.Println("starting controller")
        if !cache.WaitForCacheSync(ch, c.depCacheSynced) {//需要传递给连接的informer
//另外个if判断，如果没有这样做，有问题在等待缓存同步时将出现了问题
                fmt.Print("waiting for cache to be synced\n")
        }

        go wait.Until(c.worker, 1*time.Second, ch)
//等待直到它的作用是它在每个持续时间后调用一个特定函数，直到该特定channel关闭，所以如果
//这个函数运行我们要指定的函数，这个函数将在每个周期之后运行，直到关闭
//所以如果我们不关闭这个channel,我们将通过这个函数每次我们指定之后都会贝调用，所以我们
//称它为调用函数c.worker,指定时间秒，这样我们指定的通道就会将在这里运行，这就是一个很好的例程


        <-ch//定义始终运行，并且监听所有资源
}
```

我们必须向他传递一个`channel,`另外这个`channel`类型为控结构，所以让我们传递一个`run(ch <-chan struct{})` 

空结构`channel`的输入

现在我们需要做的是这两个函数将对象添加到数据结构到队列中，我们在图中还讨论了将要运行例程一个函数

所以让我们现在尝试实现它，

最后编写函数运行此例程

```
func (c * controller) worker () {

}

```

目前我们已经有很好的骨架，我们需要继续前进，在`main.go`

定义新的控制器，它期望客户端设置并期望部署前，所以需要部署`deploy`部署公式，像这样

```
c := newController(clientset, informers.Apps().V1().Deployments())
```

我们将调用c.run，与channel进行调用,创建channel并启动informer

```
ch := make(chan struct{})
informers.Start(ch)
```

#### 最后效果

监听所有命名空间下的deployment的变化,当然只是deploy类型
```
[root@kubecto ctoexposed]# ./ctoexposed
starting controller
add was called
add was called
add was called
add was called
add was called
```

尝试删除资源

```
del was called
```



