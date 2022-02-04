1、由于使用不同的k8s集群与之对应的`client-go`版本也不一致，拿到代码首先`go mod init app`，

进行初始化项目 

2、`go mod tidy`，解决项目依赖 

3、运行代码

  `go build -o ./list .`

4、构建镜像

```go
docker build -t list .
```

5、运行job

```go
kubectl create job list --image=list --dry-run -o yaml
```

```go
apiVersion: batch/v1
kind: Job
metadata:
  name: list
spec:
  template:
    metadata:
    spec:
      containers:
      - image: list
        imagePullPolicy: IfNotPresent
        name: list
      restartPolicy: Never
```

6、关于使用镜像运行的pod连接k8s集群时可以使用inclusterconfig,即使用令牌连接，相关源码可以在

[https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/client-go/rest/config.go#L512]


找到

```go
func InClusterConfig() (*Config, error) {
	const (
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)
```
