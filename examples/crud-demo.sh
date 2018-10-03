#!/usr/bin/env bash

default_dir=$(pwd)
tmpdir=$(mktemp -d)
project_name=odo-$RANDOM
cd $tmpdir

echo -e "\n#######################################################################"
echo -e "# Creation of Spring Boot CRUD cloud application using innerloop \n"
echo -e "# Steps :   \n"
echo -e "#  1. Create OpenShift project \n"
echo -e "#  2. Download the sd's client and install it locally \n"
echo -e "#  3. Create component yaml config containing parameters for the CRUD service \n"
echo -e "#  4. Create a Dev's supervisord pod - innerloop \n"
echo -e "#  5. Create Postgresql Database using OAB Broker \n"
echo -e "#  6. Create a secret containing the parameters to access the database and inject them within the Dev's pod \n"
echo -e "#  7. Scaffold a Spring Boot application using as template - CRUD \n"
echo -e "#  8. Compile it to create the uber jar file \n"
echo -e "#  9. Push it to the Dev's pod \n"
echo -e "# 10. Launch remotely the Spring Boot application \n"
echo -e "# 11. Curl the /api/fruit endpoint \n"
echo -e "\n####################################################"
sleep 20

echo -e "\n####################################################"
echo "# 1. Log on to the cluster and create a namespace #"
echo "####################################################"
oc login https://192.168.99.50:8443 -u admin -p admin > /dev/null 2>&1
oc new-project $project_name

echo -e "\n##########################################################"
echo "# 2. Fetch the snowdrop go client and install it locally #"
echo "##########################################################"
curl -L https://github.com/snowdrop/spring-boot-cloud-devex/releases/download/v0.15.0/sb-darwin-amd64 -o sd
chmod +x sd
export PATH=$PATH:$(pwd)

echo -e "\n###################################################################"
echo -e "\n 3. Create Component YAML file containing application's info #"
echo "###################################################################"
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

echo -e "\n#####################################"
echo "# 4. Create Dev's pod - supervisord #"
echo "#####################################"
sd init
sleep 10

echo -e "\n##############################################"
echo "# 5. Create Postgresql Service using Catalog #"
echo "##############################################"
sd catalog create my-postgresql-db -n $project_name
sleep 10

echo -e "\n#####################################################"
echo "# 6. Inject secret as ENV vars within the dev's pod #"
echo "#####################################################"
sd catalog bind --secret my-postgresql-db-secret --toInstance my-postgresql-db -n $project_name
sleep 10

echo -e "\n#######################################################################"
echo "# 7. Scaffold a Spring Boot project using CRUD template and compile it #"
echo "########################################################################"
sd create -t crud -i my-spring-boot
mvn clean package -DskipTests=true
sleep 5

echo -e "\n##############################################"
echo "# 8. Push the uber jar file to the dev's pod #"
echo "##############################################"
sd push --mode binary -n $project_name

echo -e "\n#################################################"
echo "# 9. Start the Spring Boot Application remotely #"
echo "#################################################"
sd exec start -n $project_name > /dev/null 2>&1 &

echo -e "\n######################################################################################"
echo "# 10. Curl the route of the service exposed to get the fruits using /api/fruits #"
echo "#################################################################################"
export SERVICE=$(oc get route/my-spring-boot --template='{{.spec.host}}')
for run in {1..3}
do
  sleep 5
  echo -e "\n# Call the endpoint /api/fruits"
  curl http://$SERVICE/api/fruits
  echo -e "\n"
done

echo "#######################"
echo "# 11. Clean up #"
echo "################"
cd $default_dir
rm -rf $tmpdir
oc delete project $project_name