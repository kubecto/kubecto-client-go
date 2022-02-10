
![未命名文件(5)](https://user-images.githubusercontent.com/94602819/153387551-d3ce125c-5f16-487c-b269-dadeabd3755d.png)

项目需求，通过监听`deployment`的创建，让控制器自动去创建对应的`svc`，那么自动创建以及监听的逻辑实际上都是去交给这个控制器去完成的，而上一节我们也是通过informer创建了添加删除的接口，并且实时监听deployment的状态

今天我们接着优化此项目，前面我们已经实现了实时监听`deployment`状态的添加的接口了，那么我们就可以再根据此接口去创建`svc`,凡是创建了`deployment`就可以创建对应的`svc`,来完成自动化的操作。

其中包含cluster ip以及nodeport类型，欢迎各位云原生开发者继续提供新的代码，提供其他的暴露方式

