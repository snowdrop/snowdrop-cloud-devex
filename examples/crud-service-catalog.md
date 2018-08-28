# Scaffold a Spring Boot project using a CRUD template and connect to a DB selected from a catalog

## Prerequisites

- SB tool is installed (>= 0.6.0). See README.md 
- Minishift (>= v1.23.0+91235ee) with Service Catalog feature enabled

## Step by step instructions

- Bootstrap minishift using this configuration

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
  - Pass the information and parameters needed to create the `Service Instance` and secret with the DB's paramerters  
 which supports `Postgresql`'s database

```yaml
name: my-spring-boot
env:
  - name: SPRING_PROFILES_ACTIVE
    value: openshift
service:
  - class: dh-postgresql-apb
    name: my-postgresql-db
    plan: dev
    parameter:
      - name: postgresql_user
        value: luke
      - name: postgresql_password
        value: secret
      - name: postgresql_database
        value: my_data
      - name: postgresql_version
        value: 9.6  
```

- Create a service's Instance

```bash
sb instance create -c dh-postgresql-apb
```

REMARK : The `-c` parameter will be used as key to find within the MANIFEST the service's instance config. WDYT ?

- Bind the service to generate a secret

```bash
sb bind <name_of_instance> <secret_name>
```

where : `<name_of_instance>` corresponds to the service's instance name created previously `my-postgresql-db` and `<secret_name>` is the name of the secret containing the parameters to be injected to the Development's pod.

- Initialize the Development's pod 

```bash
sb init -n crud
```

- Mount the secret as Env Vars to the Development's pod

```bash
sb mount <deployment_config_name> <secret_name>
```

where : `<deployment_config_name>` is the name of the DeploymentConfig of the Application and `<secret_name>`, the secret created previously

QUESTION: We should certyainly mount the secret during the step `sb init` to avoid to have to do it using `sb mount` command. WDYT ?


- Scaffold the CRUD project using as artifactId - `my-spring-boot` name

```bash
sb create -t crud -i my-spring-boot
```

- Generate the binary uber jar file and push it

```bash
mvn clean package
sb push --mode binary
```

- Start the Spring Boot application

```bash
sb exec start
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