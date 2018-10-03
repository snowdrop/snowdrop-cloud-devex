#!/usr/bin/env bash

default_dir=$(pwd)
tmpdir=$(mktemp -d)
project_name=odo-$RANDOM
cd $tmpdir

echo -e "\n#################################################################"
echo -e "# Creation of Spring Boot CRUD cloud application using innerloop "
echo -e "# Steps :   "
echo -e "#  1. Create OpenShift project "
echo -e "#  2. Download the sd's client and install it locally "
echo -e "#  3. Create component yaml config containing parameters for the CRUD service "
echo -e "#  4. Create a Dev's supervisord pod - innerloop "
echo -e "#  5. Create Postgresql Database using OAB Broker "
echo -e "#  6. Create a secret containing the parameters to access the database and inject them within the Dev's pod "
echo -e "#  7. Scaffold a Spring Boot application using as template - CRUD "
echo -e "#  8. Compile it to create the uber jar file "
echo -e "#  9. Push it to the Dev's pod "
echo -e "# 10. Launch remotely the Spring Boot application "
echo -e "# 11. Curl the /api/fruit endpoint "
echo -e "#################################################################"
sleep 10
clear

echo -e "\n####################################################"
echo "# 1. Log on to the cluster and create a namespace #"
echo "####################################################"
echo "oc login https://192.168.99.50:8443 -u admin -p admin"
echo "oc new-project $project_name"
sleep 5
oc login https://192.168.99.50:8443 -u admin -p admin > /dev/null 2>&1
oc new-project $project_name
clear

echo -e "\n##########################################################"
echo "# 2. Fetch the snowdrop go client and install it locally #"
echo "##########################################################"
echo "curl -L https://github.com/snowdrop/spring-boot-cloud-devex/releases/download/v0.15.0/sb-darwin-amd64 -o sd"
echo -e "chmod +x sd && export PATH=PATH:CURRENT_DIR\n"
curl -L https://github.com/snowdrop/spring-boot-cloud-devex/releases/download/v0.15.0/sb-darwin-amd64 -o sd
chmod +x sd && export PATH=$PATH:$(pwd)
clear

echo -e "\n###################################################################"
echo -e "\n 3. Create Component YAML file containing application's info #"
echo "###################################################################"
echo "cat > MANIFEST << EOF"
echo "name: my-spring-boot"
echo "env:"
echo "  - name: SPRING_PROFILES_ACTIVE"
echo "    value: openshift-catalog"
echo "services:"
echo "  - class: dh-postgresql-apb"
echo "    name: my-postgresql-db"
echo "    plan: dev"
echo "    parameters:"
echo "      - name: postgresql_user"
echo "        value: luke"
echo "      - name: postgresql_password"
echo "        value: secret"
echo "      - name: postgresql_database"
echo "        value: my_data"
echo "      - name: postgresql_version"
echo "        value: 9.6"
echo -e "EOF\n"

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
sleep 10
clear

echo -e "\n#####################################"
echo "# 4. Create Dev's pod - supervisord #"
echo "#####################################"
echo -e "sd init\n"
sd init
sleep 10
clear

echo -e "\n##############################################"
echo "# 5. Create Postgresql Service using Catalog #"
echo "##############################################"
echo -e "sd catalog create my-postgresql-db -n $project_name\n"
sd catalog create my-postgresql-db -n $project_name
sleep 10
clear

echo -e "\n#####################################################"
echo "# 6. Inject secret as ENV vars within the dev's pod #"
echo "#####################################################"
echo -e "sd catalog bind --secret my-postgresql-db-secret --toInstance my-postgresql-db -n $project_name\n"
sd catalog bind --secret my-postgresql-db-secret --toInstance my-postgresql-db -n $project_name
sleep 10
clear

echo -e "\n#######################################################################"
echo "# 7. Scaffold a Spring Boot project using CRUD template and compile it #"
echo "########################################################################"
echo -e "sd create -t crud -i my-spring-boot"
echo -e "mvn clean package -DskipTests=true\n"
sd create -t crud -i my-spring-boot
mvn clean package -DskipTests=true
sleep 5
clear

echo -e "\n##############################################"
echo "# 8. Push the uber jar file to the dev's pod #"
echo "##############################################"
echo -e "sd push --mode binary -n $project_name\n"
sd push --mode binary -n $project_name
sleep 5
clear

echo -e "\n#################################################"
echo "# 9. Start the Spring Boot Application remotely #"
echo "#################################################"
echo -e "sd exec start -n $project_name\n"
sd exec start -n $project_name > /dev/null 2>&1 &
sleep 15
clear

echo -e "\n#################################################################################"
echo "# 10. Curl the route of the service exposed to get the fruits using /api/fruits #"
echo "#################################################################################"
echo -e "export SERVICE=$(oc get route/my-spring-boot --template='{{.spec.host}}')\n"
export SERVICE=$(oc get route/my-spring-boot --template='{{.spec.host}}')
for run in {1..3}
do
  echo -e "\n# http http://$SERVICE/api/fruits"
  http http://$SERVICE/api/fruits
  echo -e "\n"
  sleep 5
done

echo "######################"
echo "# 11. Clean up #"
echo "################"
cd $default_dir
rm -rf $tmpdir
oc delete project $project_name