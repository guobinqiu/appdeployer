# 快速发布应用到kubernetes集群或者vm集群

建设中...

## TODO

- 扩展Deployment的其他常用特性, 包括滚动发布的策略, cpu和内存的资源隔离配置, 弹性伸缩等
- 目前只支持发布go,java,node,还需要包含其他语言,如python
- ansible的role还需要优化,目前go和java的应用都是nohup在后台执行, 需要配置systemctl或者supervisor等
- ansible的ssh安全性设置
- 命令行用法的描述需要优化
- 缺少使用手册(中/英), 所有的命令参数可先参见config.ini, 包括其默认值
- 存在各类bugs
- 命令行逻辑提取到API, 增加WebUI
- ...

## 发布到Kubernetes集群

不同的集群环境可以给`--kube.kubeconfig`参数设置不同的kubeconfig文件, 目前镜像用的docker, 可以配置私有镜像

```
go run main.go kube --default.appdir=~/workspace/hellogo --docker.username=qiuguobin  --kube.kubeconfig=~/Downloads/config --docker.password=*** --kube.deployment.cpulimit=500m --kube.deployment.memlimit=512mi

go run main.go kube --default.appdir=~/workspace/hellojava --docker.username=qiuguobin  --kube.kubeconfig=~/Downloads/config --docker.password=***

go run main.go kube --default.appdir=~/workspace/hellonode --docker.username=qiuguobin  --kube.kubeconfig=~/Downloads/config --docker.password=***
```

## 发布到虚拟机集群

执行命令会自动安装ansible控制机和受控机, 不同的集群环境可以给`--ansible.hosts`参数设置不同的主机列表,用逗号分隔

```
go run main.go vm --default.appdir=~/workspace/hellogo --ssh.username=guobin --ssh.password=111111 --ansible.ansible_become_password=111111 --ansible.hosts=127.0.0.1 --ansible.ansible_port=2222 --ansible.role=go

go run main.go vm --default.appdir=~/workspace/hellojava --ssh.username=guobin --ssh.password=111111 --ansible.ansible_become_password=111111 --ansible.hosts=127.0.0.1 --ansible.ansible_port=2222 --ansible.role=java

go run main.go vm --default.appdir=~/workspace/hellonode --ssh.username=guobin --ssh.password=111111 --ansible.ansible_become_password=111111 --ansible.hosts=127.0.0.1 --ansible.ansible_port=2222 --ansible.role=node
```
