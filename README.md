# Instructions to inject a Supervisord's initcontainer and enrich the Deployment of a Spring Boot S2I Application

## Download the project

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

- Execute the `go` program locally to deploy the `Java S2I - Supervisord` pod

  ```bash
  go run create.go -kubeconfig=$HOME/.kube/config
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
  
- Cleanup
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

