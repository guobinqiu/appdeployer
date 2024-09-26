# Quick Deployment of Applications to Kubernetes or VM Clusters

## Parameter Description

### Default Parameters

| Parameter | Description             | Required | Default Value  |
| --------- | ----------------------- | -------- | -------------- |
| appdir    | Local project directory | Yes      |
| appname   | Application name        | No       | Directory name |

### Git Parameters

| Parameter | Description                                                        | Required     | Default Value |
| --------- | ------------------------------------------------------------------ | ------------ | ------------- |
| enabled   | Whether to pull the project from git to the specified local appdir | No           | false         |
| repo      | Repository name                                                    | enabled=true |
| username  | Git username                                                       | enabled=true |
| password  | Git password                                                       | enabled=true |

### SSH Parameters

| Parameter             | Description                                                                                         | Required | Default Value          |
| --------------------- | --------------------------------------------------------------------------------------------------- | -------- | ---------------------- |
| username              | SSH username                                                                                        | Yes      |
| password              | SSH user password                                                                                   | Yes      |
| port                  | SSH port                                                                                            | No       | 22                     |
| authorized_keys_path  | Path to where the authorized_keys file generated on the SSH server storing SSH client's public keys | No       | ~/.ssh/authorized_keys |
| privatekey_path       | Path to where the private key file generated on the SSH client                                      | No       | ~/.ssh/appdeployer     |
| publickey_path        | Path to where the public key file generated on the SSH client                                       | No       | ~/.ssh/appdeployer.pub |
| knownhosts_path       | Path to where the known_hosts file generated on the SSH client storing SSH server's public keys     | No       | ~/.ssh/known_hosts     |
| stricthostkeychecking | SSH client verifies the server's public key before establishing a connection                        | No       | false                  |

### Ansible Parameters

| Parameter       | Description                               | Required | Default Value                                                      |
| --------------- | ----------------------------------------- | -------- | ------------------------------------------------------------------ |
| hosts           | List of remote machines                   | Yes      | localhost, separate multiple hosts with commas, supports wildcards |
| role            | Application type                          | Yes      | Supported types: go, java, nodejs                                  |
| become_password | Password for sudo execution               | Yes      |
| installdir      | Installation directory on remote machines | No       | ~/workspace                                                        |

### Docker Parameters

| Parameter    | Description                                                                                                 | Required | Default Value               |
| ------------ | ----------------------------------------------------------------------------------------------------------- | -------- | --------------------------- |
| dockerconfig | Path to Docker's config file, usually located in the user's home directory under .docker/config.json        | No       | ~/.docker/config.json       |
| dockerfile   | Path to the Dockerfile, which describes how to build the Docker image. Defaults to the root directory       | No       | ./Dockerfile                |
| registry     | URL of the Docker registry for pushing or pulling Docker images. Defaults to Docker Hub's official registry | No       | https://index.docker.io/v1/ |
| username     | Username for accessing the Docker registry. Required if the registry requires authentication                | Yes      |
| password     | Password or access token corresponding to the username. Required if the registry requires authentication    | Yes      |
| repository   | Name of the Docker image repository, including the namespace if applicable (e.g., username/repository)      | Yes      |
| tag          | Tag of the Docker image to distinguish different versions or builds in the same repository                  | No       | latest                      |

### Kube Parameters

| Parameter                                     | Description                                                                        | Required | Default Value           |
| --------------------------------------------- | ---------------------------------------------------------------------------------- | -------- | ----------------------- |
| kubeconfig                                    | Path to the Kubernetes cluster config file, used for interacting with the cluster. | No       | ~/.kube/config          |
| namespace                                     | Namespace in Kubernetes for resource isolation                                     | No       | Same as default.appname |
| ingress.host                                  | Domain or IP address for the Ingress resource to access the service                | No       | appName + ".com"        |
| ingress.tls                                   | Whether to enable TLS encryption                                                   | No       | false                   |
| ingress.selfsigned                            | Whether to use a self-signed certificate                                           | No       | false                   |
| ingress.selfsignedyears                       | Valid years for the self-signed certificate                                        | No       | 1                       |
| ingress.crtpath                               | Path to the custom TLS certificate (.crt file)                                     | No       |
| ingress.keypath                               | Path to the custom TLS key (.key file)                                             | No       |
| service.port                                  | Port number exposed by the Service                                                 | No       | 8000                    |
| deployment.replicas                           | Number of replicas in the Deployment                                               | No       | 1                       |
| deployment.port                               | Port number the application listens to inside the container                        | No       | 8000                    |
| deployment.rollingupdate.maxsurge             | Maximum number of additional replicas allowed during rolling updates               | No       | 1                       |
| deployment.rollingUpdate.maxunavailable       | Maximum number of unavailable replicas during rolling updates                      | No       | 0                       |
| deployment.quota.cpulimit                     | CPU limit for the container                                                        | No       | 1000m                   |
| deployment.quota.memlimit                     | Memory limit for the container                                                     | No       | 512Mi                   |
| deployment.quota.cpurequest                   | CPU request for the container                                                      | No       | 500m                    |
| deployment.quota.memrequest                   | Memory request for the container                                                   | No       | 256Mi                   |
| deployment.livenessprobe.enabled              | Whether to enable the liveness probe                                               | No       | false                   |
| deployment.livenessprobe.type                 | Type of liveness probe (httpget, exec, tcpsocket), case insensitive                | No       | httpget                 |
| deployment.livenessprobe.path                 | HTTP path for the liveness probe                                                   | No       | /                       |
| deployment.livenessprobe.scheme               | HTTP scheme for the liveness probe (http, https), case insensitive                 | No       | http                    |
| deployment.livenessprobe.command              | Command for the liveness probe (used when type is exec)                            | No       |
| deployment.livenessprobe.initialdelayseconds  | Initial delay in seconds for the liveness probe                                    | No       | 0                       |
| deployment.livenessprobe.timeoutseconds       | Timeout in seconds for the liveness probe                                          | No       | 1                       |
| deployment.livenessprobe.periodseconds        | Check interval in seconds for the liveness probe                                   | No       | 10                      |
| deployment.livenessprobe.successthreshold     | Success threshold for the liveness probe                                           | No       | 1                       |
| deployment.livenessprobe.failurethreshold     | Failure threshold for the liveness probe                                           | No       | 3                       |
| deployment.readinessprobe.enabled             | Whether to enable the readiness probe                                              | No       | false                   |
| deployment.readinessprobe.type                | Type of readiness probe (httpget, exec, tcpsocket), case insensitive               | No       | httpget                 |
| deployment.readinessprobe.path                | HTTP path for the readiness probe                                                  | No       | /                       |
| deployment.readinessprobe.scheme              | HTTP scheme for the readiness probe (http, https), case insensitive                | No       | http                    |
| deployment.readinessprobe.command             | Command for the readiness probe (used when type is exec)                           | No       |
| deployment.readinessprobe.initialdelayseconds | Initial delay in seconds for the readiness probe                                   | No       | 0                       |
| deployment.readinessprobe.timeoutseconds      | Timeout in seconds for the readiness probe                                         | No       | 1                       |
| deployment.readinessprobe.periodseconds       | Check interval in seconds for the readiness probe                                  | No       | 10                      |
| deployment.readinessprobe.successthreshold    | Success threshold for the readiness probe                                          | No       | 1                       |
| deployment.readinessprobe.failurethreshold    | Failure threshold for the readiness probe                                          | No       | 3                       |
| deployment.volumemount.enabled                | Whether to enable volume mount                                                     | No       | false                   |
| deployment.volumemount.mountpath              | Volume mount path                                                                  | No       | /app/data               |
| hpa.enabled                                   | Whether to enable Horizontal Pod Autoscaler                                        | No       | false                   |
| hpa.minreplicas                               | Minimum number of Pod replicas to scale down to                                    | No       | 1                       |
| hpa.maxreplicas                               | Maximum number of Pod replicas to scale up to                                      | No       | 10                      |
| hpa.cpurate=50                                | CPU utilization threshold for scaling Pod                                          | No       | 50                      |
| pvc.accessmode                                | Access mode for PVC (readwriteonce, readonlymany, readwritemany), case insensitive | No       | readwriteonce           |
| pvc.storageclassname                          | StorageClass used by the PVC                                                       | No       | openebs-hostpath        |
| pvc.storagesize                               | Requested storage size for the PVC                                                 | No       | 1Gi                     |

## Usage

### 1. Web UI 

https://github.com/guobinqiu/appdeployer-webui

### 2. CLI

Deploy to Kubernetes Cluster

```
go run main.go kube --default.appdir=~/workspace/hellogo --docker.username=qiuguobin --docker.password=*** --kube.kubeconfig=~/Downloads/config -e TZ=Asia/Shanghai

go run main.go kube --default.appdir=~/workspace/hellojava --docker.username=qiuguobin --docker.password=*** --kube.kubeconfig=~/Downloads/config -e TZ=Asia/Shanghai

go run main.go kube --default.appdir=~/workspace/hellonode --docker.username=qiuguobin --docker.password=*** --kube.kubeconfig=~/Downloads/config -e TZ=Asia/Shanghai
```

Deploy to VM Cluster

```
go run main.go vm --default.appdir=~/workspace/hellogo --ssh.username=guobin --ssh.password=111111 --ansible.become_password=111111 --ansible.hosts=192.168.1.9 --ansible.role=go

go run main.go vm --default.appdir=~/workspace/hellojava --ssh.username=guobin --ssh.password=111111 --ansible.become_password=111111 --ansible.hosts=192.168.1.9 --ansible.role=java

go run main.go vm --default.appdir=~/workspace/hellonode --ssh.username=guobin --ssh.password=111111 --ansible.become_password=111111 --ansible.hosts=192.168.1.9 --ansible.role=nodejs
```

### 3. API

Deploy to Kubernetes Cluster

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

Deploy to VM Cluster

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

## Todo List

- [ ] Add label to Deployment for traffic distribution
- [ ] Select multiple environments for simultaneous deployment
- [ ] Add static code analysis
- [ ] Add username and password login to the web UI
- [ ] Retain the history of each deployment
- [ ] ...
