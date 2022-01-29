课程示例目标：

向您展示如何使用 `client-go` 配置客户端，以从 `Kubernetes` 集群内运行的应用程序向 `Kubernetes API` 进行身份验证。

当前项目代码

```go
[root@kubecto client-go]# ls
Dockerfile  main.go  pod.yaml
```

初始化项目

```go
[root@kubecto client-go]# go mod init app
```

增加国内`goproxy`代理

```go
[root@kubecto client-go]#  export GOPROXY=https://goproxy.cn
```

自动获取依赖

```go
[root@kubecto client-go]# go mod tidy
go: finding module for package k8s.io/apimachinery/pkg/apis/meta/v1
go: finding module for package k8s.io/apimachinery/pkg/api/errors
go: finding module for package k8s.io/client-go/kubernetes
go: finding module for package k8s.io/client-go/rest
go: found k8s.io/apimachinery/pkg/api/errors in k8s.io/apimachinery v0.23.3
go: found k8s.io/apimachinery/pkg/apis/meta/v1 in k8s.io/apimachinery v0.23.3
go: found k8s.io/client-go/kubernetes in k8s.io/client-go v0.23.3
go: found k8s.io/client-go/rest in k8s.io/client-go v0.23.3
```

编译此项目

```go
[root@kubecto client-go]# go build -o ./app .
[root@kubecto client-go]# ls
app  Dockerfile  go.mod  go.sum  main.go  pod.yaml
```

然后使用提供的 `Dockerfile` 将其打包到一个 `docker` 镜像中，以便在 `Kubernetes` 上运行它。

```go
[root@kubecto client-go]# docker build -t in-cluster .
Sending build context to Docker daemon  39.66MB
Step 1/3 : FROM debian
 ---> 6f4986d78878
Step 2/3 : COPY ./app /app
 ---> 5d919827e699
Step 3/3 : ENTRYPOINT /app
 ---> Running in f9fbb624f226
Removing intermediate container f9fbb624f226
 ---> 2bf577062d65
Successfully built 2bf577062d65
Successfully tagged in-cluster:latest
[root@kubecto client-go]# docker images |grep in-cluster
in-cluster                                                        latest     2bf577062d65   13 seconds ago   163MB
```

如果您在集群上启用了 RBAC，请使用以下代码段创建角色绑定，这将授予默认服务帐户查看权限。

```go
kubectl create clusterrolebinding default-view --clusterrole=view --serviceaccount=default:default
```

在`podyaml`当中更换新的镜像，这里为`in-cluster` 然后运行

```go
[root@kubecto client-go]# kubectl apply -f pod.yaml
pod/test-incluster-client-go create
```

查看日志

```go
Every 2.0s: kubectl logs test-incluster-client-go                            Sat Jan 29 11:50:47 2022

There are 13 pods in the cluster
Found test-node-local-dns pod in default namespace
There are 13 pods in the cluster
Found test-node-local-dns pod in default namespace

```
