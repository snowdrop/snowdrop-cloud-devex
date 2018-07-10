# Instructions to inject an initcontainer containing supervisord and enrich the Deployment of a S2I Application

- Install the project
  ```bash
  cd $GOPATH/src
  go get github.com/cmoulliard/k8s-supervisor
  cd k8s-supervisor && dep ensure
  ```   

- Create `k8s-supervisord` project on OpenShift
  ```bash
  oc new-project k8s-supervisord
  ```
- Export Docker ENV var to access the docker daemon and next login
  ```bash
  eval $(minishift docker-env)
  docker login -u admin -p $(oc whoami -t) $(minishift openshift registry)
  ```
  
- Build the `copy-supervisord` docker image containing the `go supervisord` application

  **WARNING**: In order to build a multi-stages docker image, it is required to install [imagebuilder](https://github.com/openshift/imagebuilder) 

  ```bash
  cd supervisord
  imagebuilder -t $(minishift openshift registry)/k8s-supervisord/copy-supervisord:1.0 .
  ```

- Next, compile the spring Boot application using maven's tool to package the application as a `uberjar` file

  ```bash
  cd spring-boot
  mvn clean package
  rm -rf target/*-1.0.jar
  ```
  
- Build the docker image of the `Spring Boot Application` and push it to the `OpenShift` docker registry. 
 
  ```bash
  docker build -t $(minishift openshift registry)/k8s-supervisord/spring-boot-http:1.0 . -f Dockerfile
  ```   
  
- Deploy the `SpringBoot` application without the `initContainer`, shared volume, ...
  ```bash
  oc create -f openshift/spring-boot.yaml
  ```  

- Execute the go program locally to enhance the `DeploymentConfig`

  **REMARK**: Rename $HOME with the full path to access your `.kube/config` folder

  ```bash
  go run *.go -kubeconfig=$HOME/.kube/config
  Fetching about DC to be injected
  Listing deployments in namespace k8s-supervisord: 
  spring-boot-supervisord
  Updated deployment...
  ```

- Verify if the `initContainer` has been injected within the `DeploymentConfig`

  ```bash
  oc get dc/spring-boot-supervisord -o yaml | grep -A 25 initContainer
  
  initContainers:
  - args:
    - -r
    - /opt/supervisord
    - ' /var/lib/'
    command:
    - /usr/bin/cp
    image: 172.30.1.1:5000/k8s-supervisord/copy-supervisord:1.0
    imagePullPolicy: Always
    name: copy-supervisord
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/lib/supervisord
      name: shared-data
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  terminationGracePeriodSeconds: 30
  volumes:
  - emptyDir: {}
    name: shared-data
  ...
  ```

- Trigger a Deployment by pushing the images
  ```bash
  docker push $(minishift openshift registry)/k8s-supervisord/copy-supervisord:1.0
  docker push $(minishift openshift registry)/k8s-supervisord/spring-boot-http:1.0
  ```
  
- Check status of the pod created and next start a command
  ```bash
  SB_POD=$(oc get pods -l app=spring-boot-supervisord -o name)

  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl status
  echo                             STOPPED   
  run-java                         STOPPED   
  compile-java                     STOPPED   

  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl start run-java
  run-java: started
  
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl status          
  echo                             STOPPED   
  run-java                         RUNNING   pid 35, uptime 0:00:04
  compile-java                     STOPPED   
  ```  

- Let's now compile the project after pushing the project (pom, src)
  ```bash
  SB_POD=$(oc get pods -l app=spring-boot-supervisord -o name)
  SB_POD_NAME=${SB_POD:5:${#SB_POD}}
  oc cp ./pom.xml $SB_POD_NAME:/tmp/src/ -c spring-boot-supervisord
  oc cp ./src $SB_POD_NAME:/tmp/src/ -c spring-boot-supervisord
  oc rsh $SB_POD ls -la /tmp/src/
  
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl start compile-java 
  oc logs $SB_POD -f 
  
- Start the java application compiled
  ```bash
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl start run-java
  oc logs $SB_POD -f 
  ```
  
- Create an odo application and then push the project
  ```bash
  cd spring-boot
  odo app create springbootapp
  go run ../
  // TODO -> Change our code to create an application using new supervisord's approach
  odo push
  ```
    

