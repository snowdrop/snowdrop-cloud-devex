#!/usr/bin/env bash

default_dir=$(pwd)
tmpdir=$(mktemp -d)
project_name=odo-$RANDOM
cd $tmpdir

echo -e "\n#################################################################"
echo -e "# Install a Spring Boot JPA application on OpenShift which consumes a PostgeSQL's service from the catalog"
echo -e "# Steps :   "
echo -e "#  1. Log on to the cluster and create a project "
sleep 2
echo -e "#  2. Fetch the odo client and install it locally "
sleep 2
echo -e "#  3. Create a Postgresql service using the Openshift's Service Broker "
sleep 2
echo -e "#  4. Generate a Spring Boot project using a CRUD quickstart and compile it to create the uber jar file "
sleep 3
echo -e "#  5. Create a Dev's pod on OpenShift to deploy the Spring Boot application "
sleep 3
echo -e "#  6. Inject the secret as ENV vars within the dev's pod to configure the Datasource"
sleep 2
echo -e "#  7. Push it to the Dev's pod "
sleep 2
echo -e "#  8. Create a route to access the service "
sleep 2
echo -e "#  9. Access the /api/fruit endpoint (requires httpie)"
sleep 2
echo -e "#################################################################"
sleep 5
clear

echo -e "\n#################################################"
echo "# 1. Log on to the cluster and create a project #"
echo "#################################################"
echo "oc login https://192.168.99.50:8443 -u admin -p admin"
echo "oc new-project $project_name"
sleep 7
oc login https://192.168.99.50:8443 -u admin -p admin > /dev/null 2>&1
oc new-project $project_name
clear

echo -e "\n##################################################"
echo "# 2. Fetch the odo client and install it locally #"
echo "##################################################"
echo "sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.14/odo-darwin-amd64 -o /usr/local/bin/odo"
echo -e "sudo chmod +x /usr/local/bin/odo"
sudo curl -L https://github.com/redhat-developer/odo/releases/download/v0.0.14/odo-darwin-amd64 -o /usr/local/bin/odo
sudo chmod +x /usr/local/bin/odo
sleep 7
clear

echo -e "\n############################################################"
echo "# 3. Create a Postgresql Service using the Service Catalog #"
echo "############################################################"
echo -e "odo service create dh-postgresql-apb --plan dev -p postgresql_user=luke -p postgresql_password=secret -p postgresql_database=my_data -p postgresql_version=9.6\n"
odo service create dh-postgresql-apb --plan dev -p postgresql_user=luke -p postgresql_password=secret -p postgresql_database=my_data -p postgresql_version=9.6
sleep 10
clear

echo -e "\n########################################################################"
echo "# 4. Scaffold a Spring Boot project using CRUD template and compile it #"
echo "########################################################################"
echo -e "curl -o app.zip http://spring-boot-generator.195.201.87.126.nip.io/app?template=crud\n"
echo -e "unzip app.zip && rm app.zip\n"
curl -o app.zip http://spring-boot-generator.195.201.87.126.nip.io/app?template=crud
unzip app.zip && rm app.zip
sleep 7
clear

echo -e "\n######################################"
echo "# mvn clean package -DskipTests=true #"
echo "######################################"
mvn clean package -DskipTests=true
sleep 7
clear

echo -e "\n############################################################################"
echo "# 5. Create a Dev's pod on OpenShift to deploy the Spring Boot application #"
echo "############################################################################"
echo -e "odo create redhat-openjdk18-openshift:1.4 my-spring-boot --binary ./target/demo-0.0.1-SNAPSHOT.jar --env SPRING_PROFILES_ACTIVE=openshift-catalog"
odo create openjdk18:latest my-spring-boot --binary ./target/demo-0.0.1-SNAPSHOT.jar --env SPRING_PROFILES_ACTIVE=openshift-catalog
sleep 7
clear

echo -e "\n#################################################################################"
echo "# 6. Inject secret as ENV vars within the dev's pod to configure the Datasource #"
echo "#################################################################################"
echo -e "odo link dh-postgresql-apb\n"
odo link dh-postgresql-apb
sleep 10
clear

echo -e "\n##############################################"
echo "# 7. Push the uber jar file to the dev's pod #"
echo "##############################################"
echo -e "odo push\n"
odo push
sleep 7
clear

echo -e "\n###########################################################"
echo "# 8. Create a route to access the endpoint of the service #"
echo "###########################################################"
echo -e "odo url create --port 8080\n"
odo url create --port 8080
sleep 12
clear

echo -e "\n################################################################################"
echo "# 9. Curl the route of the service exposed to get the fruits using /api/fruits #"
echo "################################################################################"
echo -e "export SERVICE=$(oc get route/my-spring-boot-app --template='{{.spec.host}}')\n"
export SERVICE=$(oc get route/my-spring-boot-app --template='{{.spec.host}}')
for run in {1..3}
do
  echo -e "\n# http http://$SERVICE/api/fruits"
  http http://$SERVICE/api/fruits
  echo -e "\n"
  sleep 5
done

echo "#####################"
echo "# 10. Clean up #"
echo "################"
cd $default_dir
rm -rf $tmpdir
oc delete project $project_name