`DiscoveryClient`是发现客户端，它主要用于发现`Kubernetes API Server`所支持的资源组、资源版本、资源信息。`Kubernetes API Server`支持很多资源组、资源版本、资源信息，开发者在开发过程中很难记住所有信息，此时可以通过`DiscoveryClient`查看所支持的资源组、资源版本、资源信息。

`kubectl`的`api-versions`和`api-resources`命令输出也是通过`DiscoveryClient`实现的。另外，`DiscoveryClient`同样在`RESTClient`的基础上进行了封装。

`DiscoveryClient`除了可以发现`Kubernetes API Server`所支持的资源组、资源版本、资源信息，还可以将这些信息存储到本地，用于本地缓存（`Cache`），以减轻对`Kubernetes API Server`访问的压力。在运行`Kubernetes`组件的机器上，缓存信息默认存储于`～/.kube/cache`和`～/.kube/http-cache`下。

类似于`kubectl`命令，通过`DiscoveryClient`列出`Kubernetes API Server`所支持的资源组、资源版本、资源信息，`DiscoveryClient Example`代码示例如下：

```go
package main

import (
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func main() {

	var kubeconfig *string

	// home是家目录，如果能取得家目录的值，就可以用来做默认值
	if home:=homedir.HomeDir(); home != "" {
		// 如果输入了kubeconfig参数，该参数的值就是kubeconfig文件的绝对路径，
		// 如果没有输入kubeconfig参数，就用默认路径~/.kube/config
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		// 如果取不到当前用户的家目录，就没办法设置kubeconfig的默认目录了，只能从入参中取
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	// 从本机加载kubeconfig配置文件，因此第一个参数为空字符串
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)

	// kubeconfig加载失败就直接退出了
	if err != nil {
		panic(err.Error())
	}

	// 新建discoveryClient实例
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	// 获取所有分组和资源数据
	APIGroup, APIResourceListSlice, err := discoveryClient.ServerGroupsAndResources()

	if err != nil {
		panic(err.Error())
	}

	// 先看Group信息
	fmt.Printf("APIGroup :\n\n %v\n\n\n\n",APIGroup)

	// APIResourceListSlice是个切片，里面的每个元素代表一个GroupVersion及其资源
	for _, singleAPIResourceList := range APIResourceListSlice {

		// GroupVersion是个字符串，例如"apps/v1"
		groupVerionStr := singleAPIResourceList.GroupVersion

		// ParseGroupVersion方法将字符串转成数据结构
		gv, err := schema.ParseGroupVersion(groupVerionStr)

		if err != nil {
			panic(err.Error())
		}

		fmt.Println("*****************************************************************")
		fmt.Printf("GV string [%v]\nGV struct [%#v]\nresources :\n\n", groupVerionStr, gv)

		// APIResources字段是个切片，里面是当前GroupVersion下的所有资源
		for _, singleAPIResource := range singleAPIResourceList.APIResources {
			fmt.Printf("%v\n", singleAPIResource.Name)
		}
	}
}
```

运行以上代码，列出`Kubernetes API Server`所支持的资源组、资源版本、资源信息。首先加载`kubeconfig`配置信息，`discovery.NewDiscoveryClientForConfig`通过`kubeconfig`配置信息实例化`discoveryClient`对象，该对象是用于发现`Kubernetes API Server`所支持的资源组、资源版本、资源信息的客户端。

`discoveryClient.ServerGroupsAndResources`函数会返回`Kubernetes API Server`所支持的资源组、资源版本、资源信息（即`APIResourceList`），通过遍历`APIResourceList`输出信息。

获取`Kubernetes API Server`所支持的资源组、资源版本、资源信息
`Kubernetes API Server`暴露出`/api`和`/apis`接口。`DiscoveryClient`通过`RESTClient`分别请求`/api`和`/apis`接口，从而获取`Kubernetes API Server`所支持的资源组、资源版本、资源信息。其核心实现位于`ServerGroupsAndResources→ServerGroups`中，代码示例如下：
代码路径：`vendor/k8s.io/client-go/discovery/discovery_client.go`

```go
// ServerGroups返回支持的组，以及支持的版本和首选版本等信息。
func (d *DiscoveryClient) ServerGroups() (apiGroupList *metav1.APIGroupList, err error) {
	// 获取在/api公开的groupVersions
	v := &metav1.APIVersions{}
	err = d.restClient.Get().AbsPath(d.LegacyPrefix).Do(context.TODO()).Into(v)
	apiGroup := metav1.APIGroup{}
	if err == nil && len(v.Versions) != 0 {
		apiGroup = apiVersionsToAPIGroup(v)
	}
	if err != nil && !errors.IsNotFound(err) && !errors.IsForbidden(err) {
		return nil, err
	}

	// 获取在/api公开的groupVersions
	apiGroupList = &metav1.APIGroupList{}
	err = d.restClient.Get().AbsPath("/apis").Do(context.TODO()).Into(apiGroupList)
	if err != nil && !errors.IsNotFound(err) && !errors.IsForbidden(err) {
		return nil, err
	}
	// 为了与v1.0服务器兼容，如果它是403或404，忽略并返回我们从/api得到的内容
	if err != nil && (errors.IsNotFound(err) || errors.IsForbidden(err)) {
		apiGroupList = &metav1.APIGroupList{}
	}

	// 如果不是空的，将从/api中检索到的组前置到列表中
	if len(v.Versions) != 0 {
		apiGroupList.Groups = append([]metav1.APIGroup{apiGroup}, apiGroupList.Groups...)
	}
	return apiGroupList, nil
}
```

首先，`DiscoveryClient`通过`RESTClient`请求`/api`接口，将请求结果存放于`metav1.APIVersions`结构体中。然后，再次通过`RESTClient`请求`/apis`接口，将请求结果存放于`metav1.APIGroupList`结构体中。最后，将`/api`接口中检索到的资源组信息合并到`apiGroupList`列表中并返回。

本地缓存的`DiscoveryClient`
`DiscoveryClient`可以将资源相关信息存储于本地，默认存储位置为`～/.kube/cache`和`～/.kube/http-cache`。缓存可以减轻`client-go`对`Kubernetes API Server`的访问压力。默认每`10`分钟与`Kubernetes API Server`同步一次，同步周期较长，因为资源组、源版本、资源信息一般很少变动。本地缓存的`DiscoveryClient`

`DiscoveryClient`第一次获取资源组、资源版本、资源信息时，首先会查询本地缓存，如果数据不存在（没有命中）则请求`Kubernetes API Server`接口（回源），`Cache`将`Kubernetes API Server`响应的数据存储在本地一份并返回给`DiscoveryClient`。当下一次`DiscoveryClient`再次获取资源信息时，会将数据直接从本地缓存返回（命中）给`DiscoveryClient`。本地缓存的默认存储周期为`10`分钟。代码示例如下：

代码路径：`vendor/k8s.io/client-go/discovery/cached/disk/cached_discovery.go`

```go
// ServerResourcesForGroupVersion返回一个组和版本支持的资源。
func (d *CachedDiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	filename := filepath.Join(d.cacheDirectory, groupVersion, "serverresources.json")
	cachedBytes, err := d.getCachedFile(filename)
	//不要在错误时失败，我们要么没有一个文件，要么将无法运行缓存检查。不管怎样，我们都可以撤退。
		cachedResources := &metav1.APIResourceList{}
		if err := runtime.DecodeInto(scheme.Codecs.UniversalDecoder(), cachedBytes, cachedResources); err == nil {
			klog.V(10).Infof("returning cached discovery info from %v", filename)
			return cachedResources, nil
		}
	}

	liveResources, err := d.delegate.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		klog.V(3).Infof("skipped caching discovery info due to %v", err)
		return liveResources, err
	}
	if liveResources == nil || len(liveResources.APIResources) == 0 {
		klog.V(3).Infof("skipped caching discovery info, no resources found")
		return liveResources, err
	}

	if err := d.writeCachedFile(filename, liveResources); err != nil {
		klog.V(1).Infof("failed to write cache to %v due to %v", filename, err)
	}

	return liveResources, nil
}
```
