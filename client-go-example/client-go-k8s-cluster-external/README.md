此示例向您展示如何使用 `client-go` 配置客户端，以从在 `Kubernetes` 集群外部运行的应用程序向 `Kubernetes API` 进行身份验证。

您可以使用包含集群上下文信息的 `kubeconfig` 文件来初始化客户端。该`kubectl`命令还使用 `kubeconfig` 文件对集群进行身份验证。

当前项目代码

```go
[root@kubecto cluster-wai]# ls
go.mod  go.sum  main.go
```

运行`go`项目

```go
[root@kubecto cluster-wai]# go run main.go
There are 13 pods in the cluster
Found pod test-node-local-dns in namespace default
```
