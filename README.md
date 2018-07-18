# Cloud Native Developer's experience - Prototype

The prototype developed within this projects aims to resolve the following user's stories.

"As a user, I want to install a pod running my runtime Java Application (Spring Boot, Vert.x, Swarm) where my endpoint is exposed as a Service accessible via a route, where I can instruct the pod to start or stop a command such as "compile", "run java", ..." within the pod"

"As a user, I want to customize the application deployed using a MANIFEST yaml file where I can specify, the name of the application, s2i image to be used, maven tool, port of the service, cpu, memory, ...."

"As a user, I would like to know according to the OpenShift platform, which version of the template/builder and which resources are processed when it will be installed/deployed"

# Technical ideas

The following chapter describes how we have technically implemented such user stories:

- pod of the application/component (created by odo) defined with a :
  - initContainer : supervisord [2] where different commands are registered from ENV vars. E.g. start/stop the java runtime, debug or compile (= maven), ... 
  - container : created using Java S2I image
  - shared volume 
- commands can be executed remotely to trigger and action within the developer's pod -> supervisord ctl start|stop program1,....,programN
- OpenShift Template -> converted into individual yaml files (= builder concept) and containing "{{.key}} to be processed by the go template engine
- Developer's user preferences are stored into a MANIFEST yaml (as Cloudfoundry proposes too) which is parsed at bootstrap to create an "Application" struct object used next to process the template and replace the keys with their values

- [1] https://github.com/cmoulliard/k8s-supervisor#create-the-deploymentconfig-for-the-local-spring-boot-project
- [2] https://github.com/redhat-developer/odo/issues/556

# Table of Contents

   * [Cloud Native Developer's experience - Prototype](#cloud-native-developers-experience---prototype)
   * [Technical ideas](#technical-ideas)
   * [Table of Contents](#table-of-contents)
   * [Instructions](#instructions)
      * [Prerequisites](#prerequisites)
      * [Download the project and install it](#download-the-project-and-install-it)
      * [Create the development's pod running the supervisord](#create-the-developments-pod-running-the-supervisord)
      * [Push the code](#push-the-code)
      * [Compile the Spring Boot Java App](#compile-the-spring-boot-java-app)
      * [Start the java application](#start-the-java-application)
      * [Remote debug the Java Application](#remote-debug-the-java-application)
      * [Clean up](#clean-up)
      * [Developer section to build the images](#developer-section-to-build-the-images)

 
# Instructions

## Prerequisites

- Go Lang : [>=1.9](https://golang.org/doc/install)
- [GOWORKSPACE](https://golang.org/doc/code.html#Workspaces) variable defined
- [jq](https://stedolan.github.io/jq/)
- [oc Client Tools](https://www.openshift.org/download.html)

## Download the project and install it

- Install the project within your `$GOPATH`'s workspace
  ```bash
  cd $GOPATH/src
  go get github.com/cmoulliard/k8s-supervisor
  cd k8s-supervisor
  ```   

- Build the project locally
  ```bash
  go build -o sb *.go
  export PATH=$(pwd):$PATH
  ```
  
- Create the `k8s-supervisord` namespace on OpenShift
  ```bash
  oc new-project k8s-supervisord
  ```  

## Create the development's pod running the supervisord

- Move to the `spring-boot` folder
  ```bash
  cd spring-boot
  ```
- Execute the following `go` program locally and optionally pass as parameters:
  - `-k | --kubeconfig` : /PATH/TO/KUBE/CONFIG
  - `-n | --namespace` : openshift's project

  ```bash
  sb init
  ```

- Verify if the `initContainer` has been injected within the `DeploymentConfig`

  ```bash
  oc get dc/spring-boot-supervisord -o json | jq '.spec.template.spec.initContainers[0]'
  
  {
    "env": [
      {
        "name": "CMDS",
        "value": "echo:/var/lib/supervisord/conf/echo.sh;run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp"
      }
    ],
    "image": "quay.io/snowdrop/supervisord@sha256:0c6ad373a3aa991edcb9b5806aecd0d57467f74018c96c6299adb9a10aaa86da",
    "imagePullPolicy": "Always",
    "name": "copy-supervisord",
    "resources": {},
    "terminationMessagePath": "/dev/termination-log",
    "terminationMessagePolicy": "File",
    "volumeMounts": [
      {
        "mountPath": "/var/lib/supervisord",
        "name": "shared-data"
      }
    ]
  }
  ```
  
## Push the code  
  
- As the Development's pod has been created and is running the `supervisord` server, we will now push the code.
 
- if we want to compile the project using maven within the pod, then we will copy the following resources within the pod : `pom.xml, src/ folder`
  In this case, use the following command 

  ```bash
  sb push --mode source
  ```
  
- To use your generated `uberjar` file located under `/target/application-name-version.jar`, then run this command :
  
  ```bash
  sb push --mode binary
  ```
  
## Compile the Spring Boot Java App
  
- Compile the code source pushed within the pod using this command 
    
  ```bash
  sb compile
  INFO[0000] sb Compile command called                    
  INFO[0000] [Step 1] - Parse MANIFEST of the project if it exists 
  INFO[0000] [Step 2] - Get K8s config file               
  INFO[0000] [Step 3] - Create kube Rest config client using config's file of the developer's machine 
  INFO[0000] [Step 4] - Wait till the dev's pod is available 
  INFO[0000] [Step 5] - Compile ...                       
  compile-java: started
  time="2018-07-13T10:59:05Z" level=info msg="create process:run-java"
  time="2018-07-13T10:59:05Z" level=info msg="create process:compile-java"
  time="2018-07-13T10:59:05Z" level=info msg="create process:build"
  time="2018-07-13T10:59:05Z" level=info msg="create process:echo"
  time="2018-07-13T11:03:39Z" level=debug msg="no auth required"
  time="2018-07-13T11:03:39Z" level=debug msg="succeed to find process:compile-java"
  time="2018-07-13T11:03:39Z" level=info msg="try to start program" program=compile-java
  time="2018-07-13T11:03:39Z" level=info msg="success to start program" program=compile-java
  ==================================================================
  Starting S2I Java Build .....
  Maven build detected
  Initialising default settings /tmp/artifacts/configuration/settings.xml
  Setting MAVEN_OPTS to -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:+UseParallelOldGC -XX:MinHeapFreeRatio=10 -XX:MaxHeapFreeRatio=20 -XX:GCTimeRatio=4 -XX:AdaptiveSizePolicyWeight=90 -XX:MaxMetaspaceSize=100m -XX:+ExitOnOutOfMemoryError
  Found pom.xml ... 
  Running 'mvn -Dmaven.repo.local=/tmp/artifacts/m2 -s /tmp/artifacts/configuration/settings.xml -e -Popenshift -DskipTests -Dcom.redhat.xpaas.repo.redhatga -Dfabric8.skip=true package --batch-mode -Djava.net.preferIPv4Stack=true '
  Apache Maven 3.5.0 (Red Hat 3.5.0-4.3)
  Maven home: /opt/rh/rh-maven35/root/usr/share/maven
  Java version: 1.8.0_171, vendor: Oracle Corporation
  Java home: /usr/lib/jvm/java-1.8.0-openjdk-1.8.0.171-8.b10.el7_5.x86_64/jre
  Default locale: en_US, platform encoding: ANSI_X3.4-1968
  OS name: "linux", version: "3.10.0-693.21.1.el7.x86_64", arch: "amd64", family: "unix"
  time="2018-07-13T11:03:40Z" level=debug msg="wait program exit" program=compile-java
  [INFO] Error stacktraces are turned on.
  [INFO] Scanning for projects...
  [INFO] Downloading: https://repo1.maven.org/maven2/io/openshift/booster-parent/23/booster-parent-23.pom
  ...
  [INFO] ------------------------------------------------------------------------
  [INFO] BUILD SUCCESS
  [INFO] ------------------------------------------------------------------------
  [INFO] Total time: 01:07 min
  [INFO] Finished at: 2018-07-13T11:04:49Z
  [INFO] Final Memory: 26M/40M
  [INFO] ------------------------------------------------------------------------
  [WARNING] The requested profile "openshift" could not be activated because it does not exist.
  Copying Maven artifacts from /tmp/src/target to /deployments ...
  Running: cp *.jar /deployments
  ... done
  time="2018-07-13T11:04:49Z" level=info msg="program stopped with status:exit status 0" program=compile-java
  time="2018-07-13T11:04:49Z" level=info msg="Don't start the stopped program because its autorestart flag is false" program=compile-java
  ```
  
  **Remark** Before to launch the compilation's command using supervisord, the program will wait till the development's pod is alive !
  
## Start the java application

- Launch the Spring Boot Application

  ```bash
  sb exec start
  ime="2018-07-13T11:06:26Z" level=debug msg="succeed to find process:run-java"
  time="2018-07-13T11:06:26Z" level=info msg="try to start program" program=run-java
  time="2018-07-13T11:06:26Z" level=info msg="success to start program" program=run-java
  Starting the Java application using /opt/run-java/run-java.sh ...
  exec java -javaagent:/opt/jolokia/jolokia.jar=config=/opt/jolokia/etc/jolokia.properties -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:+UseParallelOldGC -XX:MinHeapFreeRatio=10 -XX:MaxHeapFreeRatio=20 -XX:GCTimeRatio=4 -XX:AdaptiveSizePolicyWeight=90 -XX:MaxMetaspaceSize=100m -XX:+ExitOnOutOfMemoryError -cp . -jar /deployments/spring-boot-http-1.0.jar
  time="2018-07-13T11:06:27Z" level=debug msg="wait program exit" program=run-java
  I> No access restrictor found, access to any MBean is allowed
  Jolokia: Agent started with URL https://172.17.0.7:8778/jolokia/
    .   ____          _            __ _ _
   /\\ / ___'_ __ _ _(_)_ __  __ _ \ \ \ \
  ( ( )\___ | '_ | '_| | '_ \/ _` | \ \ \ \
   \\/  ___)| |_)| | | | | || (_| |  ) ) ) )
    '  |____| .__|_| |_|_| |_\__, | / / / /
   =========|_|==============|___/=/_/_/_/
   :: Spring Boot ::       (v1.5.14.RELEASE)
  ... 
  2018-07-13 11:06:34.293  INFO 222 --- [           main] o.s.j.e.a.AnnotationMBeanExporter        : Registering beans for JMX exposure on startup
  2018-07-13 11:06:34.304  INFO 222 --- [           main] o.s.c.support.DefaultLifecycleProcessor  : Starting beans in phase 0
  2018-07-13 11:06:34.427  INFO 222 --- [           main] s.b.c.e.t.TomcatEmbeddedServletContainer : Tomcat started on port(s): 8080 (http)
  2018-07-13 11:06:34.436  INFO 222 --- [           main] io.openshift.booster.BoosterApplication  : Started BoosterApplication in 6.412 seconds (JVM running for 7.32) 
  ```
  
- Access the endpoint of the Spring Boot application using curl
  ```bash
  URL="http://$(oc get routes/spring-boot-http -o jsonpath='{.spec.host}')"
  curl $URL/api/greeting
  {"content":"Hello, World!"}% 
  ``` 
  
## Remote debug the Java Application
  
- You can also debug your application by forwarding the traffic between the pod and your machine using the following command : 
  ```bash
  sb debug
  INFO[0000] sb exec start command called                        
  INFO[0000] [Step 1] - Parse MANIFEST of the project if it exists 
  INFO[0000] [Step 2] - Get K8s config file               
  INFO[0000] [Step 3] - Create kube Rest config client using config's file of the developer's machine 
  INFO[0000] [Step 4] - Wait till the dev's pod is available 
  INFO[0000] [Step 5] - Restart Java Application          
  run-java: stopped
  run-java: started
  INFO[0003] [Step 6] - Remote Debug the spring Boot Application ... 
  Forwarding from 127.0.0.1:5005 -> 5005
  ```
  
  **Remark** : You can change the local/remote ports to be used by passing the parameter `-p`. E.g `sb debug -p 9009:9009`
  
## Stop/start or restart the spring boot application

- The Spring Boot Application can be stopped, started or restarted using respectively these commands:
  ```bash
  sb exec stop
  sb exec start
  sb exec restart
  ```  
  
## Clean up
  
  ```bash
  oc delete --force --grace-period=0 all --all
  oc delete pvc/m2-data
  ``` 

