# k8s-supervisor

- Compile locally

```bash
mkdir ~/Temp/go-supervisor
export GOPATH=~/Temp/go-supervisor
go get -u github.com/ochinchina/supervisord
```

- Create k8s-supervisor project

```bash
oc new-project k8s-supervisor
eval $(minishift docker-env)
```
- Build docker image and push it

```bash
docker login -u admin -p $(oc whoami -t) $(minishift openshift registry)
imagebuilder -t $(minishift openshift registry)/k8s-supervisor/supervisord:v1 -f Dockerfile .
docker push $(minishift openshift registry)/k8s-supervisor/supervisord:v1
```

- Create application using `oc new-app` command

```bash
oc new-app --name=docker-supervisord -i supervisord:v1 -l app=supervisord
--> Found image 236a45c (51 seconds old) in image stream "k8s-supervisor/supervisord" under tag "v1" for "supervisord:v1"

    * This image will be deployed in deployment config "docker-supervisord"
    * The image does not expose any ports - if you want to load balance or send traffic to this component
      you will need to create a service with 'expose dc/docker-supervisord --port=[port]' later
    * WARNING: Image "k8s-supervisor/supervisord:v1" runs as the 'root' user which may not be permitted by your cluster administrator

--> Creating resources with label app=supervisord ...
    imagestreamtag "docker-supervisord:v1" created
    deploymentconfig "docker-supervisord" created
--> Success
    Run 'oc status' to view your app.
```

- Create application using pod file

```bash
oc delete -f supervisord-pod.yml
oc create -f supervisord-pod.yml
oc logs supervisord-pod
```

- Create supervisord conf

```bash
cat > supervisor.conf << 'EOF'
cat supervisor.conf
[program:test]
command = /your/program args
[inet_http_server]
port=127.0.0.1:9001
EOF

```

- all in one

```bash
imagebuilder -t $(minishift openshift registry)/k8s-supervisor/supervisord:v1 -f Dockerfile .
docker push $(minishift openshift registry)/k8s-supervisor/supervisord:v1
oc delete -f supervisord-pod.yml
sleep 10s
oc create -f supervisord-pod.yml
oc logs supervisord-pod
```











- Test Initcontainer

```bash
cat > src/my-app.yml << 'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: myapp-pod
  labels:
    app: myapp
spec:
  containers:
  - name: myapp-container
    image: busybox
    command: ['sh', '-c', 'echo The app is running! && sleep 3600']
  initContainers:
  - name: init-myservice
    image: busybox
    command: ['sh', '-c', 'until nslookup myservice; do echo waiting for myservice; sleep 2; done;']
  - name: init-mydb
    image: busybox
    command: ['sh', '-c', 'until nslookup mydb; do echo waiting for mydb; sleep 2; done;']
EOF

cat > src/service.yml << 'EOF'
kind: Service
apiVersion: v1
metadata:
  name: myservice
spec:
  ports:
  - protocol: TCP
    port: 80
    targetPort: 9376
---
kind: Service
apiVersion: v1
metadata:
  name: mydb
spec:
  ports:
  - protocol: TCP
    port: 80
    targetPort: 9377
EOF

```

- Install the pod

```bash
oc create -f src/my-app.yml
...
oc logs myapp-pod -c init-myservice
oc logs myapp-pod -c init-mydb
oc create -f src/service.yml
oc get -f src/myapp.yaml
```
