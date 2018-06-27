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
- Build the `supervisord` docker image using our configuration file and push it

```bash
docker login -u admin -p $(oc whoami -t) $(minishift openshift registry)
imagebuilder -t $(minishift openshift registry)/k8s-supervisor/supervisord:1.0 -f Dockerfile-supervisord .
docker push $(minishift openshift registry)/k8s-supervisor/supervisord:1.0
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

# Spring Boot Application and initContainer

- We use maven to package the project as a `uberjar` file

  ```bash
  mvn clean package
  rm -rf target/*-1.0.jar
  ```
  
- Build the docker image of the Spring Boot Application, like also the supervisord and push it to the OpenShift docker registry
 
  ```bash
  docker build -t $(minishift openshift registry)/k8s-supervisor/spring-boot-http:1.0 . -f Dockerfile-spring-boot
  docker push $(minishift openshift registry)/k8s-supervisor/spring-boot-http:1.0
  docker build -t $(minishift openshift registry)/k8s-supervisor/copy-supervisord:1.0 -f Dockerfile-copy-supervisord .
  docker push $(minishift openshift registry)/k8s-supervisor/copy-supervisord:1.0
  ```  
  
- The application is created on the cloud platform
  ```bash
  oc create -f openshift/spring-boot-supervisord.yaml
  docker push $(minishift openshift registry)/k8s-supervisor/copy-supervisord:1.0
  docker push $(minishift openshift registry)/k8s-supervisor/spring-boot-http:1.0
  ```  