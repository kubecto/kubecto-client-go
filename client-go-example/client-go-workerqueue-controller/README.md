
### Informer

`Informer` 负责观察资源的目标状态，这个过程需要是可伸缩和可持续性的。它们还负责实现重新同步机制 ，从而保证定时地解决冲突。它们通常用于保持集群的实际狀态与內存中的缓存的状态一致（代码 Bug 或是网络问题可能会导致它们不一致）

### WorkQueue

事件处理器把状态变化情况放入工作队列，用于保证在必要时可以进行重试。在 `Client-go` 中，这个功能是通过 `workqueue` 包来提供的。当在对资源或外部系统进行变更或者更新状态时发生错误，资源可以被重新放回工作队列，它也可以被用于其他不能实时处理的状态变化的情况，把相关资源放入队列后便可以在稍晚些时候再次处理。



### 使用client-go当中的workqueue和informer实现一个无限循环的控制器

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
