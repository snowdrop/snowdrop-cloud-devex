# Scaffold a Spring Boot application using JPA's module and connect it to a Database deployed as a service

## Prerequisites

- Minishift (>= v1.23.0+91235ee) with Service Catalog feature enabled

## Install tools

- odo which will provide the ability to run the spring boot application on Openshift and also create the PostgreSQL database

To build odo from source, perform the following commands (requires golang tools to be setup properly)

```bash
cd $GOPATH/src/github.com/redhat-developer
git clone https://github.com/redhat-developer/odo.git && cd odo
make install && sudo cp $GOPATH/bin/odo /usr/local/bin
```

Alternatively you can install the latest odo release (requires you have [jq](https://stedolan.github.io/jq/) setup)

```bash
# if you already have a minishift VM that was not started like below, it needs to be deleted with minishift delete
minishift addons enable xpaas
minishift addons enable admin-user
MINISHIFT_ENABLE_EXPERIMENTAL=y minishift start --extra-clusterup-flags="--enable=service-catalog,automation-service-broker" 
# subsequent starts of the VM can be done simply with minishift start

curl -L -o odo $(curl -sL https://api.github.com/repos/redhat-developer/odo/releases/latest | jq -r '.assets[].browser_download_url' | grep 'odo-linux-amd64$') # use odo-darwin-64 for Mac
chmod +x odo
sudo cp odo /usr/local/bin
```

## Steps to follow to play the scenario

```bash
oc new-project odo

cd /Temp/my-spring-boot
# rm -rf {src,target,MANIFEST} && rm -rf *.{iml,xml,zip}
# Scaffold a JPA Persistence Spring Boot Project
curl -o app.zip http://spring-boot-generator.195.201.87.126.nip.io/app?template=crud
unzip app.zip
rm app.zip
mvn clean package

# Create the component with allow odo to run the application on Openshift
odo create redhat-openjdk18-openshift:1.3 my-spring-boot --binary /Temp/my-spring-boot/target/my-spring-boot-0.0.1-SNAPSHOT.jar --env SPRING_PROFILES_ACTIVE=openshift-catalog

# Create a service's instance using the OABroker and postgresql DB + secret. Next bind/link the secret to the DC and restart it
odo service create dh-postgresql-apb --plan dev -p postgresql_user=luke -p postgresql_password=secret -p postgresql_database=my_data -p postgresql_version=9.6

# At this point you need to wait until PostgreSQL is deployed

# Make sure that secret containing PostgreSQL connection data is passed to the container that will run the spring boot application
odo link dh-postgresql-apb

# The previous command will force a redeployment of the component

# Push the uber jar file
odo push

# Make the application accessible from outside the cluster on port 8080
odo url create --port 8080
```

After a few seconds, the application will be running on port 8080

- Query the service

```bash
APP_BASE_PATH=http://$(oc get route my-spring-boot-app -o jsonpath='{.spec.host}')
http $(APP_BASE_PATH)/api/fruits
HTTP/1.1 200 
Cache-control: private
Content-Type: application/json;charset=UTF-8
Date: Wed, 12 Sep 2018 13:57:47 GMT
Set-Cookie: d63dbdcbd8cb849d5e746267aaa4ed8d=ab5897d9975f4716398ac33b10f3c13d; path=/; HttpOnly
Transfer-Encoding: chunked
X-Application-Context: application:openshift-catalog

[
    {
        "id": 1,
        "name": "Cherry"
    },
    {
        "id": 2,
        "name": "Apple"
    },
    {
        "id": 3,
        "name": "Banana"
    }
]
```


