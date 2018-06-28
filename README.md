# Go odo controller

- Install the project

```bash
cd $GOPATH
go get 
```

- Install the dc without initContainer
```bash
oc new-project k8s-supervisord
oc create -f deploy/openshift/dc.yml
```

- Run the controller locally

```bash
go run *.go -kubeconfig=/Users/dabou/.kube/config
```