# 快速发布应用到k8s集群或者vm

建设中...

Deploy to Kubernetes cluster

```
go run main.go kube --default.appdir=~/workspace/hellogo --docker.username=qiuguobin  --kube.kubeconfig=~/Downloads/config
```

Deploy to VM set

```
go run main.go vm --default.appdir=~/workspace/hellojava --ssh.username=guobin --ssh.password=******** --ansible.ansible_become_password=111111 --ansible.hosts=127.0.0.1 --ansible.ansible_port=2222 --ansible.role=java
```
