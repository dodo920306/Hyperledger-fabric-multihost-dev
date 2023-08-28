#!/bin/bash

echo -e "\033[0;32m$(<logo.txt )\033[0m"
while true;
do
    echo -n "Enter your organization's name: "
    read ORG
    if [[ ! $ORG =~ ^[a-zA-Z0-9]+$ ]]; then
      echo "The name shouldn't include characters that are not letters and numbers."
      continue
    fi
    echo -n -e "The name you enter is \e[1;31m${ORG}\e[0m. This name can't be changed once order set up. Can you confirm that this is correct? (y/n)"
    read confirm
    if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
      break
    fi
done
while true;
do
    echo -n "Enter your organization's hostname: "
    read HOST
    if [[ ! $HOST =~ ^[a-zA-Z0-9.]+$ ]]; then
      echo "The hostname shouldn't include characters that are not letters, dots and numbers."
      continue
    fi
    echo -n -e "The hostname you enter is \e[1;31m${HOST}\e[0m. This name can't be changed once order set up. Can you confirm that this is correct? (y/n)"
    read confirm
    if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
      break
    fi
done

echo "Welcome ${ORG}!"
set -x

curl -sSLO https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh && chmod +x install-fabric.sh
sudo ./install-fabric.sh d b
rm -f install-fabric.sh
sudo cp bin/* /usr/local/bin/

sed "s/\${HOST}/${HOST}/g" template/docker-compose-ca-template.yaml > docker-compose-ca.yaml
sed "s/\${ORG}/${ORG}/" template/.env.ca.template > .env.ca
sed "s/\${HOST}/${HOST}/g" template/docker-compose-template.yaml > docker-compose.yaml
sed -e "s/\${ORG}/${ORG}/g" -e "s/\${HOST}/${HOST}/g" template/.env.orderer.template > .env.orderer
sed -e "s/\${ORG}/${ORG}/g" -e "s/\${HOST}/${HOST}/g" template/.env.peer.template > .env.peer
cp template/.env.couchdb.template .env.couchdb

sudo docker volume create orderer0
sudo docker volume create peer0

sudo chown -R `whoami`:`whoami` .

sed -e "s/\${ORG}/${ORG}/g" -e "s/\${HOST}/${HOST}/g" -e "s/\${PWD}/$(echo "${PWD}" | sed 's/\//\\\//g')/g" template/configtx-template.yaml > config/configtx.yaml

sed -e "s/\${ORG}/${ORG}/g" -e "s/\${HOST}/${HOST}/g" template/setup.sh.template > setup.sh

sed -e "s/\${ORG}/${ORG}/g" -e "s/\${HOST}/${HOST}/g" template/createChannel.sh.template > createChannel.sh

chmod a+x setup.sh
chmod a+x createChannel.sh

set +x

echo "Initiation complete."
