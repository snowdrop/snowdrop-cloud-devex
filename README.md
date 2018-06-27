# k8s-supervisor

- Compile locally and test it

```bash
mkdir ~/Temp/go-supervisor
export GOPATH=~/Temp/go-supervisor
go get -u github.com/ochinchina/supervisord
$GOPATH/bin/supervisord -c supervisor-local.conf
```

- Create k8s-supervisor project on OpenShift

```bash
oc new-project k8s-supervisor
eval $(minishift docker-env)
```
- Build the supervisord docker image and push it

```bash
docker login -u admin -p $(oc whoami -t) $(minishift openshift registry)
imagebuilder -t $(minishift openshift registry)/k8s-supervisor/supervisord:v1 -f Dockerfile-supervisord .
docker push $(minishift openshift registry)/k8s-supervisor/supervisord:v1
```

- Create a `supervisord` application using `oc new-app` command

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

- Create `supervisord` application using `pod` yaml file

```bash
oc delete -f openshift/supervisord-pod.yml
oc create -f openshift/supervisord-pod.yml
oc logs supervisord-pod
```

- All in one

```bash
imagebuilder -t $(minishift openshift registry)/k8s-supervisor/supervisord:v1 -f Dockerfile .
docker push $(minishift openshift registry)/k8s-supervisor/supervisord:v1
oc delete -f supervisord-pod.yml
sleep 10s
oc create -f supervisord-pod.yml
oc logs supervisord-pod
```
