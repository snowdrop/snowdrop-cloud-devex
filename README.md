# Cloud Native Developer's experience - Prototype

The prototype developed within this projects aims to resolve the following user's stories.

"As a user, I want to install a pod running my runtime Java Application (Spring Boot, Vert.x, Swarm), where I can instruct the tool to start or stop a command such as "compile", "run java", ..." within the pod"

"As a user, I want to customize the application deployed using a MANIFEST yaml file where I can specify, the name of the application, s2i image to be used, maven tool, port of the service, cpu, memory, ...."

"As a user, I would like to know according to the OpenShift platform, which version of the template/builder and which resources are processed when it will be installed/deployed"

# Technical's idea

The following chapter describes how we have technically implemented such user's story :

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
   * [Technical's idea](#technicals-idea)
   * [Table of Contents](#table-of-contents)
   * [Instructions](#instructions)
      * [Download the project and install it](#download-the-project-and-install-it)
      * [Create the resources on OpenShift to compile/run your Spring Boot application](#create-the-resources-on-openshift-to-compilerun-your-spring-boot-application)
      * [Push the code](#push-the-code)
      * [Compile the Spring Boot Java App](#compile-the-spring-boot-java-app)
      * [Start the java application and curl the endpoint](#start-the-java-application-and-curl-the-endpoint)
      * [Clean up](#clean-up)
      * [Developer section to build the images](#developer-section-to-build-the-images)
 
# Instructions

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
  export PATH=$PATH:$(pwd)
  ```
  
- Create the `k8s-supervisord` namespace on OpenShift
  ```bash
  oc new-project k8s-supervisord
  ```  

## Create the resources on OpenShift to compile/run your Spring Boot application

- Move to the `spring-boot` folder
  ```bash
  cd spring-boot
  ```
- Execute the following `go` program locally and pass as parameter :
  - `-k | --kubeconfig` : /PATH/TO/KUBE/CONFIG
  - `-n | --namespace` : openshift's project

  ```bash
  sb init -n k8s-supervisord
  ... 
  ```

- Verify if the `initContainer` has been injected within the `DeploymentConfig`

  ```bash
  oc get dc/spring-boot-supervisord -o yaml | grep -A 25 initContainer
  
  initContainers:
      - env:
        - name: CMDS
          value: echo:/var/lib/supervisord/conf/echo.sh;run-java:/usr/local/s2i/run;compile-java:/usr/local/s2i/assemble;build:/deployments/buildapp
        image: quay.io/snowdrop/supervisord@sha256:0c6ad373a3aa991edcb9b5806aecd0d57467f74018c96c6299adb9a10aaa86da
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
  triggers:
  - imageChangeParams:
      automatic: true
      containerNames:
      - copy-supervisord
  ...
  ```
  
## Push the code  
  
- As the Development's pod has been created and is running the `supervisord's application`, we will now push the local's code (pom.xml, src)
  to the pod

  ```bash
  sb push
  INFO[0000] sb Push command called                       
  INFO[0000] [Step 1] - Parse MANIFEST of the project if it exists 
  INFO[0000] [Step 2] - Get K8s config file               
  INFO[0000] [Step 3] - Create kube Rest config client using config's file of the developer's machine 
  INFO[0000] [Step 4] - Wait till the dev's pod is available 
  INFO[0000] [Step 5] - Copy files from Development projects to the pod 
  ```
   
  
## Compile the Spring Boot Java App
  
- Compile it within the pod  
    
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
  
## Start the java application and curl the endpoint

- Launch the Spring Boot Application

  ```bash
  sb run
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
  
- Access the endpoint of the Spring Boot application 
  ```bash
  URL="http://$(oc get routes/spring-boot-supervisord -o jsonpath='{.spec.host}')"
  curl $URL/api/greeting
  {"content":"Hello, World!"}% 
  ``` 
  
- You can also debug your application using the default port defined which is `5005` 
  ```bash
  sb debug OR sb debug -p 9009:9009
  INFO[0000] sb Run command called                        
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
  
## Clean up
  
  ```bash
  oc delete all --all
  ```  
    
## Developer section to build the images

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
  TAG_ID=$(docker images -q cmoulliard/copy-supervisord:latest)
  docker tag $TAG_ID quay.io/snowdrop/supervisord
  docker login quai.io
  docker push quay.io/snowdrop/supervisord
  ```
  
- Build the docker image of `Spring Boot S2I`
 
  ```bash
  docker build -t <username>/spring-boot-http:latest .
  docker tag 00c6b955c3e1 quay.io/snowdrop/spring-boot-s2i
  docker push quay.io/snowdrop/spring-boot-s2i
  ```    

