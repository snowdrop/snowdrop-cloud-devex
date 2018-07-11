# Cloud Native Developer's experience - Prototype

The prototype developed within this projects aims to resolve the following user's stories.

"As a user, I want to install a pod running my runtime Java Application (Spring Boot, Vert.x, Thorntail), where I can instruct the devtool (= odo) to start or stop a command such as "compile", "run java", ..." within the pod"

"As a user, I want to customize the application deployed using a MANIFEST yaml file where I can specify, the name of the application, s2i image to be used, maven tool, port of the service, cpu, memory, ...."

"As a user, I would like to know according to the OpenShift platform, which version of the template and which resources are processed when it will be installed/deployed"

List of technical features implemented are :

- pod of the application/component (created by odo) defined with a :
  - initContainer : supervisord [2] where different commands are registered from ENV vars. E.g. start/stop the java runtime, debug or compile (= maven), ... 
  - container : created using Java S2I image
  - shared volume 
- commands can be executed remotely to trigger and action within the developer's pod -> supervisord ctl start|stop program1,....,programN
- OpenShift Template -> converted into individual yaml files (= builder concept) and containing "{{.key}} to be processed by the go template engine
- Developer's user preferences are stored into a MANIFEST yaml (as Cloudfoundry proposes too) which is parsed at bootstrap [3] to create an "Application" struct object used next to process the template and replace the keys with their values [4]

[1] https://github.com/cmoulliard/k8s-supervisor#create-the-deploymentconfig-for-the-local-spring-boot-project
[2] https://github.com/redhat-developer/odo/issues/556
[3] https://goo.gl/J1bQ4x
[4] https://goo.gl/hmKdnh

# Table of Contents

   * [Cloud Native Developer's experience - Prototype](#cloud-native-developers-experience---prototype)
   * [Table of Contents](#table-of-contents)
   * [Instructions](#instructions)
      * [Download the project and install it](#download-the-project-and-install-it)
      * [Create the deploymentConfig for the local spring Boot project](#create-the-deploymentconfig-for-the-local-spring-boot-project)
      * [Push the code](#push-the-code)
      * [Compile and start the Spring Boot Java App](#compile-and-start-the-spring-boot-java-app)
      * [Clean up](#clean-up)
      * [Developer section](#developer-section)
 
# Instructions

## Download the project and install it

- Install the project within your `$GOPATH`'s workspace
  ```bash
  cd $GOPATH/src
  go get github.com/cmoulliard/k8s-supervisor
  cd k8s-supervisor && dep ensure
  ```   

- Create the `k8s-supervisord` namespace on OpenShift
  ```bash
  oc new-project k8s-supervisord
  ```  

## Create the deploymentConfig for the local spring Boot project

- Execute the `go` program locally to deploy the `Java S2I - Supervisord` pod and pass as parameters :
  - `-kubeconfig=PATH_TO_KUBE_CONFIG` 
  - Path to access the `MANIFEST` file of the Application

  ```bash
  go run create.go -kubeconfig=$HOME/.kube/config spring-boot/MANIFEST
  INFO[0000] [Step 1] - Create Kube Client & Clientset    
  INFO[0000] [Step 2] - Create ImageStreams for Supervisord and Java S2I Image of SpringBoot 
  INFO[0000] [Step 3] - Create DeploymentConfig using Supervisord and Java S2I Image of SpringBoot 
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
  
## Push the code  
  
- Check the status of the supervisord to verify the programs which are available
  ```bash
  SB_POD=$(oc get pods -l app=spring-boot-supervisord -o name)

  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl status
  echo                             STOPPED   
  run-java                         STOPPED   
  compile-java                     STOPPED   
  ```  

- Push the code source (pom.xml, ./src)
  ```bash
  SB_POD=$(oc get pods -l app=spring-boot-supervisord -o name)
  SB_POD_NAME=${SB_POD:5:${#SB_POD}}
  oc cp ./pom.xml $SB_POD_NAME:/tmp/src/ -c spring-boot-supervisord
  oc cp ./src $SB_POD_NAME:/tmp/src/ -c spring-boot-supervisord
  oc rsh $SB_POD ls -la /tmp/src/
  
## Compile and start the Spring Boot Java App
  
- Compile it within the pod  
  ```bash
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl start compile-java 
  oc logs $SB_POD -f 
  ```
  
- Start the java application compiled
  ```bash
  oc rsh $SB_POD /var/lib/supervisord/bin/supervisord ctl start run-java
  oc logs $SB_POD -f 
  ```
  
- Access the endpoint of the Spring Boot application 
  ```bash
  URL="http://$(oc get routes/spring-boot-supervisord -o jsonpath='{.spec.host}')"
  curl $URL/api/greeting
  {"content":"Hello, World!"}% 
  ``` 
  
## Clean up
  
  ```bash
  oc delete all --all
  ```  
    
## Developer section

- Export Docker ENV var to access the docker daemon
  ```bash
  eval $(minishift docker-env)
  ```

- To build the `copy-supervisord` docker image containing the `go supervisord` application, then follow these instructions

  **WARNING**: In order to build a multi-stages docker image, it is required to install [imagebuilder](https://github.com/openshift/imagebuilder) 

  ```bash
  cd supervisord
  imagebuilder -t <username>/copy-supervisord:latest .
  ```
  
- Tag the docker image and push it to `quay.io`

  ```bash
  docker tag b74c32ba6bd8 quay.io/snowdrop/supervisord
  docker login quai.io
  docker push quay.io/snowdrop/supervisord
  ```
  
- Build the docker image of `Spring Boot S2I`
 
  ```bash
  docker build -t <username>/spring-boot-http:latest .
  docker tag 00c6b955c3e1 quay.io/snowdrop/spring-boot-s2i
  docker push quay.io/snowdrop/spring-boot-s2i
  ```    

