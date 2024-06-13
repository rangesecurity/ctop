#!/bin/bash

# Make this script fail on any error
set -e

if [ "$#" -lt 1 ]; then
    echo "Parameters missing. Expected: deploy.sh <host>"
    exit 1
fi

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
host="$1"

# Change into the directory above this present script
cd "$dir"; cd ..

echo "********************************************************************************"
echo "* Build pvtop"
echo "********************************************************************************"
# tbd

echo ""
echo ""
echo "********************************************************************************"
echo "* Deploy current working tree to host..."
echo "********************************************************************************"
echo ""
export ANSIBLE_CONFIG="./ansible/ansible.cfg"
ansible-playbook --inventory-file ansible/inventories/hosts.yml --limit $host ansible/deploy.yml

