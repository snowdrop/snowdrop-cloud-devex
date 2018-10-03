#!/usr/bin/env bash

default_dir=$(pwd)
tmpdir=$(mktemp -d)
project_name=odo-$RANDOM
cd $tmpdir

echo -e "\n####################################################"
echo "# 1. Log on tpo the cluster and create a namespace #"
echo "####################################################"
oc login https://192.168.99.50:8443 -u admin -p admin
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

echo "######################################################################################"
echo "# 10. Curl the route of the service exposed to get the fruits using /api/fruits #"
echo "#################################################################################"
export SERVICE=$(oc get route/my-spring-boot --template='{{.spec.host}}')
for run in {1..5}
do
  sleep 5
  echo -e "\n# Call the endpoint /api/fruits"
  curl -s http://$SERVICE/api/fruits | json
  echo -e "\n"
done

echo "#######################"
echo "# 11. Clean up #"
echo "################"
cd $default_dir
rm -rf $tmpdir
oc delete project $project_name