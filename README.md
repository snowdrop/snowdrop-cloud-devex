# Go odo controller

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

- Install the `DeploymentConfig` of the SpringBoot application without the `initContainer`
  ```bash
  oc create -f openshift/dc.yml
  ```

- Build the `copy-supervisord` docker image containing the `go supervisord` application

  ```bash
  cd supervisord
  docker build -t $(minishift openshift registry)/k8s-supervisord/copy-supervisord:1.0 -f Dockerfile-copy-supervisord .
  docker push $(minishift openshift registry)/k8s-supervisord/copy-supervisord:1.0
  ```

- Compile the spring Boot application using maven to package the project as a `uberjar` file

  ```bash
  cd spring-boot
  mvn clean package
  rm -rf target/*-1.0.jar
  ```
  
- Build the docker image of the `Spring Boot Application` and push it to the `OpenShift` docker registry. 
 
  ```bash
  docker build -t $(minishift openshift registry)/k8s-supervisord/spring-boot-http:1.0 . -f Dockerfile-spring-boot
  docker push $(minishift openshift registry)/k8s-supervisord/spring-boot-http:1.0
  ```   

- Execute the go program locally to inject the `initContainer`

  **REMARK**: Rename $HOME with the full path to access your `.kube/config` folder

  ```bash
  go run *.go -kubeconfig=$HOME/.kube/config
  Fetching about DC to be injected
  Listing deployments in namespace k8s-supervisord: 
  spring-boot-supervisord
  Updated deployment...
  ```

- Verify if the `initContainer` has been injected with the `DeploymentConfig`

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