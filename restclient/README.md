`RESTClient`是最基础的客户端。其他的`ClientSet`、`DynamicClient`及`DiscoveryClient`都是基于`RESTClient`实现的。`RESTClient`对`HTTP Request`进行了封装，实现了`RESTful`风格的`API`。它具有很高的灵活性，数据不依赖于方法和资源，因此`RESTClient`能够处理多种类型的调用，返回不同的数据格式。

类似于`kubectl`命令，通过`RESTClient`列出`kube-system`运行的`Pod`资源对象，`RESTClient Example`代码示例：

[https://github.com/kubecto/kubecto-client-go](https://github.com/kubecto/kubecto-client-go)

```go

package main

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// 参考path : /api/v1/namespaces/{namespace}/pods
	config.APIPath = "api"
	// pod的group是空字符串
	config.GroupVersion = &corev1.SchemeGroupVersion
	// 指定序列化工具
	config.NegotiatedSerializer = scheme.Codecs

	// 根据配置信息构建restClient实例
	restClient, err := rest.RESTClientFor(config)

	if err!=nil {
		panic(err.Error())
	}

	// 保存pod结果的数据结构实例
	result := &corev1.PodList{}

	//  指定namespace
	namespace := "kube-system"
	// 设置请求参数，然后发起请求
	// GET请求
	err = restClient.Get().
		//  指定namespace，参考path : /api/v1/namespaces/{namespace}/pods
		Namespace(namespace).
		// 查找多个pod，参考path : /api/v1/namespaces/{namespace}/pods
		Resource("pods").
		// 指定大小限制和序列化工具
		VersionedParams(&metav1.ListOptions{Limit:100}, scheme.ParameterCodec).
		// 请求
		Do(context.TODO()).
		// 结果存入result
		Into(result)

	if err != nil {
		panic(err.Error())
	}

	// 打印名称
	fmt.Printf("Namespace\t Status\t\t Name\n")

	// 每个pod都打印Namespace、Status.Phase、Name三个字段
	for _, d := range result.Items {
		fmt.Printf("%v\t %v\t %v\n",
			d.Namespace,
			d.Status.Phase,
			d.Name)
	}
}
```

运行以上代码，列出`kube-system`命名空间下的所有`Pod`资源对象的相关信息。首先加载`kubeconfig`配置信息，并设置`config.APIPath`请求的`HTTP`路径。然后设置`config.GroupVersion`请求的资源组/资源版本。最后设置`config.NegotiatedSerializer`数据的编解码器。

`rest.RESTClientFor`函数通过`kubeconfig`配置信息实例化`RESTClient`对象，`RESTClient`对象构建`HTTP`请求参数，例如`Get`函数设置请求方法为`get`操作，它还支持`Post、Put、Delete、Patch`等请求方法。`Namespace`函数设置请求的命名空间。`Resource`函数设置请求的资源名称。`VersionedParams`函数将一些查询选项（如`limit、TimeoutSeconds`等）添加到请求参数中。通过`Do`函数执行该请求，并将`kube-apiserver`返回的结果（`Result`对象）解析到`corev1.PodList`对象中。最终格式化输出结果。

`RESTClient`发送请求的过程对`Go`语言标准库`net/http`进行了封装，由`Do→request`函数实现，代码示例如下：
代码路径：`vendor/k8s.io/client-go/rest/request.go`

```go
func (r *Request) Do(ctx context.Context) Result {
	var result Result
	err := r.request(ctx, func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(resp, req)
	})
	if err != nil {
		return Result{err: err}
	}
	return result
}
```

```go
for {

		url := r.URL().String()
		req, err := http.NewRequest(r.verb, url, r.body)
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)
		req.Header = r.headers

		r.backoff.Sleep(r.backoff.CalculateBackoff(r.URL()))
		if retries > 0 {
			// We are retrying the request that we already send to apiserver
			// at least once before.
			// This request should also be throttled with the client-internal rate limiter.
			if err := r.tryThrottleWithInfo(ctx, retryInfo); err != nil {
				return err
			}
			retryInfo = ""
		}
		resp, err := client.Do(req)
		updateURLMetrics(ctx, r, resp, err)
		if err != nil {
			r.backoff.UpdateBackoff(r.URL(), err, 0)
		} else {
			r.backoff.UpdateBackoff(r.URL(), err, resp.StatusCode)
		}
		if err != nil {
			// "Connection reset by peer" or "apiserver is shutting down" are usually a transient errors.
			// Thus in case of "GET" operations, we simply retry it.
			// We are not automatically retrying "write" operations, as
			// they are not idempotent.
			if r.verb != "GET" {
				return err
			}
			// For connection errors and apiserver shutdown errors retry.
			if net.IsConnectionReset(err) || net.IsProbableEOF(err) {
				// For the purpose of retry, we set the artificial "retry-after" response.
				// TODO: Should we clean the original response if it exists?
				resp = &http.Response{
					StatusCode: http.StatusInternalServerError,
					Header:     http.Header{"Retry-After": []string{"1"}},
					Body:       ioutil.NopCloser(bytes.NewReader([]byte{})),
				}
			} else {
				return err
			}
		}

		done := func() bool {
			// Ensure the response body is fully read and closed
			// before we reconnect, so that we reuse the same TCP
			// connection.
			defer func() {
				const maxBodySlurpSize = 2 << 10
				if resp.ContentLength <= maxBodySlurpSize {
					io.Copy(ioutil.Discard, &io.LimitedReader{R: resp.Body, N: maxBodySlurpSize})
				}
				resp.Body.Close()
			}()

			retries++
			if seconds, wait := checkWait(resp); wait && retries <= r.maxRetries {
				retryInfo = getRetryReason(retries, seconds, resp, err)
				if seeker, ok := r.body.(io.Seeker); ok && r.body != nil {
					_, err := seeker.Seek(0, 0)
					if err != nil {
						klog.V(4).Infof("Could not retry request, can't Seek() back to beginning of body for %T", r.body)
						fn(req, resp)
						return true
					}
				}

				klog.V(4).Infof("Got a Retry-After %ds response for attempt %d to %v", seconds, retries, url)
				r.backoff.Sleep(time.Duration(seconds) * time.Second)
				return false
			}
			fn(req, resp)
			return true
		}()
```

请求发送之前需要根据请求参数生成请求的`RESTful URL`，由`r.URL.String`函数完成。例如，在`RESTClient Example`代码示例中，根据请求参数生成请求的`RESTful URL`为`http://127.0.0.1:8080/api/v1/namespaces/kube-system/pods？limit=500`，其中`api`参数为`v1`，`namespace`参数为`system`，请求的资源为`pods`，`limit`参数表示最多检索出`500`条信息。

最后通过`Go`语言标准库`net/http`向`RESTful URL`（即`kube-apiserver`）发送请求，请求得到的结果存放在`http.Response`的`Body`对象中，`fn`函数（即`transformResponse`）将结果转换为资源对象。当函数退出时，会通过`resp.Body.Close`命令进行关闭，防止内存溢出。
