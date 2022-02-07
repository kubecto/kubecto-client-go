使用client-go实现一个无限循环的控制器

主函数main将以命令行方式传递kubeconfig文件文件

使用方法

```
# go run main.go -kubeconfig=/root/.kube/config
I0207 16:48:43.900771   15500 main.go:104] Starting Pod controller
Sync/Add/Update for Pod nginx-6799fc88d8-jlbzc
Sync/Add/Update for Pod test-job-z9vq5
Sync/Add/Update for Pod test-node-local-dns
Sync/Add/Update for Pod jobs-v8phv
Sync/Add/Update for Pod test-node-local-dns
```

删除pod，和创建pod则都会获取得到相关信息，并记录日志当中


```
删除指定pod则提示此pod not exist,也就是控制器的循环状态，以期望的状态运行
Sync/Add/Update for Pod test-job-z9vq5
Pod default/test-job-z9vq5 does not exist anymore
```

创建pod，则提示add/update for pod xxx

```
Sync/Add/Update for Pod test-nod
Sync/Add/Update for Pod test-nod
```
