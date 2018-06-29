# Instructions to use go supervisord with a spring boot application

- Git clone the project

  ```bash
  git clone https://github.com/cmoulliard/k8s-supervisor.git 
  cd k8s-supervisor.
  ```

- Compile locally the go `supervisord` and test it using our conf

  ```bash
  mkdir ~/Temp/go-supervisor
  export GOPATH=~/Temp/go-supervisor
  go get -u github.com/ochinchina/supervisord
  $GOPATH/bin/supervisord -c docker/conf/supervisor-local.conf
  ```

- Create `k8s-supervisord` project on OpenShift

  ```bash
  oc new-project k8s-supervisord
  eval $(minishift docker-env)
  docker login -u admin -p $(oc whoami -t) $(minishift openshift registry)
  ```
- Build the `copy-supervisord` docker image using our configuration file and push it

  ```bash
  cd supervisord
  docker build -t $(minishift openshift registry)/k8s-supervisord/copy-supervisord:1.0 -f Dockerfile-copy-supervisord .
  docker push $(minishift openshift registry)/k8s-supervisord/copy-supervisord:1.0
  ```

# Spring Boot Application and initContainer

- We use maven to package the project as a `uberjar` file

  ```bash
  cd spring-boot
  mvn clean package
  rm -rf target/*-1.0.jar
  ```
  
- Build the docker image of the Spring Boot Application` and push it to the OpenShift docker registry
 
  ```bash
  docker build -t $(minishift openshift registry)/k8s-supervisord/spring-boot-http:1.0 . -f Dockerfile-spring-boot
  docker push $(minishift openshift registry)/k8s-supervisord/spring-boot-http:1.0
  ```  
  
- Deploy the application on `OpenShift`
  ```bash
  oc delete all --all   
  oc create -f openshift/spring-boot-supervisord.yaml
  docker push $(minishift openshift registry)/k8s-supervisord/copy-supervisord:1.0
  docker push $(minishift openshift registry)/k8s-supervisord/spring-boot-http:1.0
  ```  
  
- Check status and test
  ```bash
  SB_POD=$(oc get pods -l app=spring-boot-supervisord -o name)
  SERVICE_IP=$(minishift ip)
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl status
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl pid run-java
  http http://sb-k8s-supervisord.$SERVICE_IP.nip.io/api/greeting
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl stop run-java
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl start run-java
  http http://sb-k8s-supervisord.$SERVICE_IP.nip.io/api/greeting
  ```  