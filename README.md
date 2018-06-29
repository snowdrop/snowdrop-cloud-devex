# Go odo controller

- Install the project

```bash
cd $GOPATH/src
go get github.com/cmoulliard/k8s-supervisor
dep ensure
```

- Install the `DeploymentConfig` of the SpringBoot application without the `initContainer`
```bash
oc new-project k8s-supervisord
oc create -f openshift/dc.yml
```

- Execute the program locally

**REMARK**: Rename $HOME with the full path to access your `.kube/config` folder

```bash
go run *.go -kubeconfig=$HOME/.kube/config
Fetching about DC to be injected
Listing deployments in namespace k8s-supervisord: 
spring-boot-supervisord
Updated deployment...
```

- Verify if the `initContainer` has been injected

```bash
oc get dc/spring-boot-supervisord -o yaml | grep -A 25 initContainer

    initContainers:
    - args:
      - -r
      - /opt/supervisord
      - ' /var/lib/'
      command:
      - /usr/bin/cp
      image: docker/dd/dd
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
test: false
triggers:
- type: ConfigChange
```