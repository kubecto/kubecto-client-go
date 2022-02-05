我们将构建一个简单的命令行工具，它将job名称、容器镜像名称和执行命令作为参数，并在 Kubernetes 集群上创建一个job。我们将使用 golang 的内置包 -`flag`来做到这一点。我们将在模块的根目录中创建一个名为的文件`main.go`，它会完成所需的一切。


```
package main

import (
    "flag"
    "fmt"
)
func main() {
    jobName := flag.String("jobname", "test-job", "The name of the job")
    containerImage := flag.String("image", "ubuntu:latest", "Name of the container image")
    entryCommand := flag.String("command", "ls", "The command to run inside the container")

    flag.Parse()

    fmt.Printf("Args : %s %s %s\n", *jobName, *containerImage, *entryCommand)
}
```

所以现在我们有了一个基本的`main`函数设置，它将简单地接受参数`jobName`，`containerName`并`imageName`打印它们。现在让我们开始使用`client-go`库。


1、编译项目完成后

```
[root@kubecto client-go-create-job-command]# ./main -h
Usage of ./main:
  -command string
    	执行容器的命令 (default "ls")
  -image string
    	容器镜像的名字 (default "ubuntu:latest")
  -jobname string
    	这是job的名字 (default "test-job")
```

2、使用go run main.go则默认执行命令

```
./main --jobname=test-job --image=ubuntu:latestt --command="ls"
```

3、当然也可以自己进行传递参数来创建新的job
