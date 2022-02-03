1、由于使用不同的k8s集群与之对应的client-go版本也不一致，拿到代码首先go mod init app，进行初始化项目

2、go mod tidy，解决项目依赖

3、运行代码
```
go run main.go
```

4、执行结果如下
```
[root@kubecto listpod]# go run node.go
获取default namespace下的pod
nginx-6799fc88d8-jlbzc
获取default namespace下的deployment的名字
nginx
获取kube-system namespace下的daemonset的名字
kube-proxy
获取get node的方法
kubecto
```

此代码使用开发技巧，首先开发list功能需要注意的点

1)、https://github.com/kubernetes/kubernetes/blob/master/pkg/registry/
通过这里找到对应的 kubernetes Group以及 kubernetes Kind/resources简称GR/GK

2)、https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/api/apps/
比如开发apps资源类型下的pod,先找到register.go，找到对应的list接口
```
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Deployment{},
		&DeploymentList{},
		&StatefulSet{},
		&StatefulSetList{},
		&DaemonSet{},
		&DaemonSetList{},
		&ReplicaSet{},
		&ReplicaSetList{},
		&ControllerRevision{},
		&ControllerRevisionList{},
```
    
3)、再去看type.go里面找到DeploymentList的结构体
```
// DeploymentList is a list of Deployments.
type DeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of Deployments.
	Items []Deployment `json:"items" protobuf:"bytes,2,rep,name=items"`
```

4)、取方法clientset.AppsV1().Deployments("default").List(ctx, metav1.ListOptions{})

5)、最后使用遍历，取值，其他方法大同小异
```
for _,d deploy range {
            fmt.Printf("%s\n", d.Name)
            }
```


  
    
```
