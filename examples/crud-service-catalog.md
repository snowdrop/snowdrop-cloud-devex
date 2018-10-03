# Scaffold a Spring Boot project using a CRUD template and connect to a DB installed using a Service Broker and Catalog

## Prerequisites

- SD tool is installed (>= [0.15.0](https://github.com/snowdrop/spring-boot-cloud-devex/releases/tag/v0.15.0)). See README.md 
- Minishift (>= v1.23.0+91235ee) with Service Catalog feature enabled

## Step by step instructions

- Bootstrap minishift using this configuration or `oc cluster up --enable=*,service-catalog,automation-service-broker`

```bash
minishift config set memory 6G
minishift config set openshift-version v3.10.0
minishift addons enable admin-user

MINISHIFT_ENABLE_EXPERIMENTAL=y minishift start --extra-clusterup-flags="--enable=*,service-catalog,automation-service-broker"
```

- Log to your OpenShift's cluster and create a crud namespace

```bash
oc login -u admin -p admin
oc new-project crud-catalog
```

- Create a `my-spring-boot` project directory and move under this directory

```bash
cd /path/to/my-spring-boot
```

- Create a `MANIFEST`'s project file within the current folder containing the :
  - ENV vars used by the spring's boot application to use the Service provisioned from the catalog
  - Pass the information and parameters needed to create the `Service Instance` and the DB's parameters  
    which supports `Postgresql`'s database

```bash
cat > MANIFEST << EOF
name: my-spring-boot
env:
  - name: SPRING_PROFILES_ACTIVE
    value: openshift-catalog
services:
  - class: dh-postgresql-apb
    name: my-postgresql-db
    plan: dev
    parameters:
      - name: postgresql_user
        value: luke
      - name: postgresql_password
        value: secret
      - name: postgresql_database
        value: my_data
      - name: postgresql_version
        value: 9.6
EOF
```

- Scaffold the CRUD project using as artifactId the `my-spring-boot` name specified within the MANIFEST 

```bash
sd create -t crud -i my-spring-boot
```

- Create a service's instance using our service instance name `my-postgresql-db`

```bash
sd catalog create <name_of_the_service_instance>
```

where `<name_of_the_service_instance>` is the name to be defined for th service that we will create using the command (e.g my-postgresql-apb).
The Service class to be selected from the catalog is specified within the MANIFEST using the field services/class `db-postgresql-apb` 

- Create a secret using the service parameters and bind/mount them to the DeploymentConfig that `sd` automatically created for you

```bash
sd catalog bind --secret <secret_name> --toInstance <name_of_the_service_instance>
```

where : `<name_of_the_service_instance>` corresponds to the service's instance name created previously `my-postgresql-db` and `<secret_name>` is the name of the secret (e.g my-postgresql-db-secret).

- Generate the Spring Boot binary uber jar file and push it to the development's pod

```bash
mvn clean package
sd push --mode binary
```

- Start the Spring Boot application

```bash
sd exec start
```

- Use `curl` or `httpie` tool to fetch the records using the Spring Boot CRUD endpoint exposed

```bash
http http://MY_APP_NAME-MY_PROJECT_NAME.OPENSHIFT_HOSTNAME/api/fruits
HTTP/1.1 200 
Cache-control: private
Content-Type: application/json;charset=UTF-8
Date: Mon, 27 Aug 2018 07:54:21 GMT
Set-Cookie: 23678da2d4b6649bf39522f45e3064f1=45e36b4277a6cd39f15bb8efaa87c882; path=/; HttpOnly
Transfer-Encoding: chunked
X-Application-Context: application:openshift

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
        "id": 3,git pull
        "name": "Banana"
    }
]
```

- Add a `fruit`

```bash
curl -H "Content-Type: application/json" -X POST -d '{"name":"pear"}' http://MY_APP_NAME-MY_PROJECT_NAME.OPENSHIFT_HOSTNAME/api/fruits
```

## Asciinema recording

- Record the video using the demo script
```bash
asciinema rec -c './examples/crud-demo.sh' crud-demo.cast
```

- Play it locally

```bash
asciinema play crud-demo.cast
```

