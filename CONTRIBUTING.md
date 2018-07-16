# Scenario to be validated

## Test 0 : Build executable and test it

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `k8s-supervisord`
- Move under the `spring-boot` folder and run these commands

```bash
go build -o sb *.go
export PATH=$PATH:$(pwd)
export CURRENT=$(pwd)

cd spring-boot
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

sb init -n k8s-supervisord
sb push --mode source
cd $CURRENT
```

## Test 1 : source -> compile -> run

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `k8s-supervisord`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go run ../main.go init -n k8s-supervisord
go run ../main.go push --mode source
go run ../main.go compile
go run ../main.go run
```

- Execute this command within another terminal

```bash
URL="http://$(oc get routes/spring-boot-http -o jsonpath='{.spec.host}')"
curl $URL/api/greeting
```

## Test 2 : binary -> run

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `k8s-supervisord`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go run ../main.go init -n k8s-supervisord
go run ../main.go push --mode binary
go run ../main.go run
```

- Execute this command within another terminal

```bash
URL="http://$(oc get routes/spring-boot-http -o jsonpath='{.spec.host}')"
curl $URL/api/greeting
```

## Test 3 : Debug

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `k8s-supervisord`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go run ../main.go init -n k8s-supervisord
go run ../main.go push --mode binary
go run ../main.go debug
```

## Test4 : source -> compile -> kill pod -> compile again

- Log on to an OpenShift cluster with an `admin` role
- Open or create the following project : `k8s-supervisord`
- Move under the `spring-boot` folder and run these commands

```bash
oc delete --force --grace-period=0 all --all
oc delete pvc/m2-data

go run ../main.go init -n k8s-supervisord
go run ../main.go push --mode source
go run ../main.go compile
oc delete --grace-period=0 --force=true pod -l app=spring-boot-http 
go run ../main.go push --mode source
go run ../main.go compile
go run ../main.go run
```