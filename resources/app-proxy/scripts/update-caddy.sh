#!/bin/bash
#
#

# change directory to our source code ( directory up from our script )
SCRIPT_DIR=$(dirname "${BASH_SOURCE[0]}")
cd ${SCRIPT_DIR}/..

#
echo "Checking for caddy update"

# run our generic github updater
./scripts/poll-github-releases.py -r caddyserver/caddy --filterregex '^v([2\.7\.[\d+]+)' -f caddy-release.txt

# pull in our version number and update our SHA256 sum
CADDY_VERSION=$(cat caddy-release.txt | tr -d '\n') 

echo "Updating our build to caddy version ${CADDY_VERSION}"

sed -i "s/CADDY_VERSION :=.*/CADDY_VERSION := ${CADDY_VERSION}/g" Makefile
