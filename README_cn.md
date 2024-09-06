# 快速发布应用到kubernetes集群或者vm集群

## 参数说明

### default 参数

| 参数名  | 参数描述         | 必填 | 默认值     |
| ------- | ---------------- | ---- | ---------- |
| appdir  | 本地项目目录路径 | 是   |
| appname | 应用名称         | 否   | 项目目录名 |

### git参数

| 参数名   | 参数描述                                        | 必填               | 默认值 |
| -------- | ----------------------------------------------- | ------------------ | ------ |
| enabled  | 是否需要从git拉取项目到本地appdir指定的目录路径 | 否                 | false  |
| repo     | 代码库名称                                      | enabled=true时必填 |
| username | git用户名                                       | enabled=true时必填 |
| password | git密码                                         | enabled=true时必填 |

### ssh参数

| 参数名                | 参数描述                                               | 必填 | 默认值                 |
| --------------------- | ------------------------------------------------------ | ---- | ---------------------- |
| username              | ssh用户名                                              | 是   |
| password              | ssh用户密码                                            | 是   |
| port                  | ssh端口                                                | 否   | 22                     |
| authorized_keys_path  | ssh服务端authorized_keys文件路径,用于存储ssh客户端公钥 | 否   | ~/.ssh/authorized_keys |
| privatekey_path       | ssh客户端私钥文件路径                                  | 否   | ~/.ssh/appdeployer     |
| publickey_path        | ssh客户端公钥文件路径                                  | 否   | ~/.ssh/appdeployer.pub |
| knownhosts_path       | ssh客户端known_hosts文件路径                           | No   | ~/.ssh/known_hosts     |
| stricthostkeychecking | ssh客户端在接受连接前对服务端公钥进行验证              | No   | true                   |

### ansible参数

| 参数名          | 参数描述                 | 必填 | 默认值                                |
| --------------- | ------------------------ | ---- | ------------------------------------- |
| hosts           | 远程机列表               | 是   | localhost,多主机用逗号分隔,支持通配符 |
| role            | 应用类型                 | 是   | 目前支持的类型:go,java,nodejs         |
| become_password | 执行sudo的密码           | 是   |
| installdir      | 应用在远程机上的安装目录 | 否   | ~/workspace                           |

### docker参数

| 参数名       | 参数描述                                                                                                                   | 必填 | 默认值                      |
| ------------ | -------------------------------------------------------------------------------------------------------------------------- | ---- | --------------------------- |
| dockerconfig | Docker的配置文件路径,通常包含认证信息和Docker仓库的访问设置.这个文件在大多数情况下位于用户的家目录下的.docker/config.json. | 否   | ~/.docker/config.json       |
| dockerfile   | Dockerfile的路径,Dockerfile是用于描述如何构建Docker镜像的文本文件.默认位于当前目录的根路径下.                              | 否   | ./Dockerfile                |
| registry     | Docker仓库的URL,用于推送或拉取Docker镜像.默认是Docker Hub的官方仓库地址.                                                   | 否   | https://index.docker.io/v1/ |
| username     | 用于访问Docker仓库的用户名.如果仓库需要认证,则此参数是必需的.                                                              | 是   |
| password     | 与username对应的密码或访问令牌.如果仓库需要认证,则此参数是必需的                                                           | 是   |
| repository   | Docker镜像的仓库名称,包括可能的命名空间（例如,username/repository）                                                        | 是   |
| tag          | Docker镜像的标签,用于区分同一仓库中的不同版本或构建                                                                        | 否   | latest                      |

### kube参数

| 参数名                                        | 参数描述                                                                                           | 必填  | 默认值            |
| --------------------------------------------- | -------------------------------------------------------------------------------------------------- | ----- | ----------------- |
| kubeconfig                                    | Kubernetes集群的配置文件路径,用于与集群进行交互.该文件包含了集群的访问权限和API服务器的地址等信息. | 否    | ~/.kube/config    |
| namespace                                     | Kubernetes中的命名空间,用于隔离资源                                                                | 否    | 同default.appname |
| ingress.host                                  | Ingress资源的域名或IP地址,用于访问服务                                                             | 否    | appName + ”.com“  |
| ingress.tls                                   | 是否启用TLS加密.否                                                                                 | false |
| ingress.selfsigned                            | 是否使用自签名证书                                                                                 | 否    | false             |
| ingress.selfsignedyears                       | 自签名证书的有效年数                                                                               | 否    | 1                 |
| ingress.crtpath                               | 自定义TLS证书的路径（.crt文件）                                                                    | 否    |
| ingress.keypath                               | 自定义TLS密钥的路径（.key文件）                                                                    | 否    |
| service.port                                  | Service暴露的端口号                                                                                | 否    | 8000              |
| deployment.replicas	Deployment的副本数量      | 否                                                                                                 | 1     |
| deployment.port                               | 容器内应用程序监听的端口号                                                                         | 否    | 8000              |
| deployment.rollingupdate.maxsurge             | 滚动更新时,允许的最大额外副本数                                                                    | 否    | 1                 |
| deployment.rollingUpdate.maxunavailable       | 滚动更新时,允许的最大不可用副本数                                                                  | 否    | 0                 |
| deployment.quota.cpulimit                     | 容器CPU使用的限制                                                                                  | 否    | 1000m             |
| deployment.quota.memlimit                     | 容器内存使用的限制                                                                                 | 否    | 512Mi             |
| deployment.quota.cpurequest                   | 容器CPU使用的请求值                                                                                | 否    | 500m              |
| deployment.quota.memrequest                   | 容器内存使用的请求值                                                                               | 否    | 256Mi             |
| deployment.livenessprobe.enabled              | 是否启用存活探针                                                                                   | 否    | false             |
| deployment.livenessprobe.type                 | 存活探针的类型(httpget,exec,tcpsocket),不区分大小写                                                | 否    | httpget           |
| deployment.livenessprobe.path                 | 存活探针的HTTP路径                                                                                 | 否    | /                 |
| deployment.livenessprobe.scheme               | 存活探针的HTTP模式(http,https),不区分大小写                                                        | 否    | http              |
| deployment.livenessprobe.command              | 存活探针的命令（当type为exec时使用）                                                               | 否    |
| deployment.livenessprobe.initialdelayseconds  | 存活探针的初始延迟秒数                                                                             | 否    | 0                 |
| deployment.livenessprobe.timeoutseconds       | 存活探针的超时秒数                                                                                 | 否    | 1                 |
| deployment.livenessprobe.periodseconds        | 存活探针的检查间隔秒数                                                                             | 否    | 10                |
| deployment.livenessprobe.successthreshold     | 存活探针的成功阈值                                                                                 | 否    | 1                 |
| deployment.livenessprobe.failurethreshold     | 存活探针的失败阈值                                                                                 | 否    | 3                 |
| deployment.readinessprobe.enabled             | 是否启用就绪探针                                                                                   | 否    | false             |
| deployment.readinessprobe.type                | 就绪探针的类型(httpget,exec,tcpsocket),不区分大小写                                                | 否    | httpget           |
| deployment.readinessprobe.path                | 就绪探针的HTTP路径                                                                                 | 否    | /                 |
| deployment.readinessprobe.scheme              | 就绪探针的HTTP模式(http,https),不区分大小写                                                        | 否    | http              |
| deployment.readinessprobe.command             | 就绪探针的命令(当type为exec时使用)                                                                 | 否    |
| deployment.readinessprobe.initialdelayseconds | 就绪探针的初始延迟秒数                                                                             | 否    | 0                 |
| deployment.readinessprobe.timeoutseconds      | 就绪探针的超时秒数                                                                                 | 否    | 1                 |
| deployment.readinessprobe.periodseconds       | 就绪探针的检查间隔秒数                                                                             | 否    | 10                |
| deployment.readinessprobe.successthreshold    | 就绪探针的成功阈值                                                                                 | 否    | 1                 |
| deployment.readinessprobe.failurethreshold    | 就绪探针的失败阈值                                                                                 | 否    | 3                 |
| deployment.volumemount.enabled                | 是否启用卷挂载                                                                                     | 否    | false             |
| deployment.volumemount.mountpath              | 卷挂载路径                                                                                         | 否    | /app/data         |
| hpa.enabled                                   | 是否启用Horizontal Pod Autoscaler                                                                  | 否    | false             |
| hpa.minreplicas                               | HPA缩小的最小Pod副本数                                                                             | 否    | 1                 |
| hpa.maxreplicas                               | HPA扩展的最大Pod副本数                                                                             | 否    | 10                |
| hpa.cpurate=50                                | HPA扩展Pod的CPU利用率阈值                                                                          | 否    | 50                |
| pvc.accessmode                                | PVC的访问模式(readwriteonce,readonlymany,readwritemany),不区分大小写                               | 否    | readwriteonce     |
| pvc.storageclassname                          | PVC所使用的StorageClass                                                                            | 否    | openebs-hostpath  |
| pvc.storagesize                               | PVC请求的存储大小                                                                                  | 否    | 1G                |

## 用法

### 发布到Kubernetes集群

不同的集群环境可以给`--kube.kubeconfig`参数设置不同的kubeconfig文件, 目前镜像用的docker, 可以配置私有镜像

#### CLI

```
go run main.go kube --default.appdir=~/workspace/hellogo --docker.username=qiuguobin --docker.password=*** --kube.kubeconfig=~/Downloads/config -e TZ=Asia/Shanghai

go run main.go kube --default.appdir=~/workspace/hellojava --docker.username=qiuguobin --docker.password=*** --kube.kubeconfig=~/Downloads/config -e TZ=Asia/Shanghai

go run main.go kube --default.appdir=~/workspace/hellonode --docker.username=qiuguobin --docker.password=*** --kube.kubeconfig=~/Downloads/config -e TZ=Asia/Shanghai
```

#### API

```
curl --location 'http://localhost:8888/kube/submit' \
--header 'Content-Type: application/json' \
--data '{
    "default": {
        "appdir": "~/workspace/hellogo"
    },
    "docker": {
        "username": "qiuguobin",
        "password": "111111111"
    }
}'

curl -X GET 'http://localhost:8888/kube/deploy?requestID=XXXXXXXXXXX'
```

SSE URL
```
http://localhost:8888/kube/deploy?requestID=1725602754229208000
```

### 发布到虚拟机集群

安装ansible,不同的集群环境可以给`--ansible.hosts`参数设置不同的主机列表,用逗号分隔

#### CLI

```
go run main.go vm --default.appdir=~/workspace/hellogo --ssh.username=guobin --ssh.password=111111 --ansible.become_password=111111 --ansible.hosts=192.168.1.9 --ansible.role=go

go run main.go vm --default.appdir=~/workspace/hellojava --ssh.username=guobin --ssh.password=111111 --ansible.become_password=111111 --ansible.hosts=192.168.1.9 --ansible.role=java

go run main.go vm --default.appdir=~/workspace/hellonode --ssh.username=guobin --ssh.password=111111 --ansible.become_password=111111 --ansible.hosts=192.168.1.9 --ansible.role=nodejs
```

#### API

```
curl --location 'http://localhost:8888/vm/submit' \
--header 'Content-Type: application/json' \
--data '{
    "default": {
        "appdir": "~/workspace/hellogo"
    },
    "ssh": {
        "username": "guobin",
        "password": "111111"
    },
    "ansible": {
        "role": "go",
        "become_password": "111111",
        "hosts": "192.168.1.9"
    }
}'

curl -X GET 'http://localhost:8888/vm/deploy?requestID=XXXXXXXXXXX'
```
