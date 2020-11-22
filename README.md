# OIP Daemon

> OIPd Monitors the FLO Blockchain, indexing OIP messages into a searchable index

[![License](https://img.shields.io/github/license/oipwg/oip.svg)](https://github.com/oipwg/oip/blob/master/LICENSE.md) [![Image Pulls](https://img.shields.io/docker/pulls/mediciland/oip.svg)](https://hub.docker.com/r/mediciland/oip) [![Image Stars](https://img.shields.io/docker/stars/mediciland/oip.svg)](https://hub.docker.com/r/mediciland/oip)

# Build
Give execute access to script first:  
`chmod +x ./build.sh`

`./build.sh`

# Run
Give execute access to script first:  
`chmod +x ./run.sh`

Then run program with either of the following commands  
`./run.sh` or `LOG=* go run ./cmd/oipd/*.go`

or after **Build**:  
`LOG=* ./bin/oipd`

# Config
`Mac`: `~/Library/Application\ /Support/Oipd`

Else, probably: `~/.oipd`

You'll need to set the flod `rpcuser` and `rpcpass` in the config.  
Look at the flod config (either in Application Support for mac or `~/flod`) to see what the rpc settings are and copy them to the oipd config.

# Docker Image
The included Docker image runs the following software in tandem to allow access to a fully functional OIP stack.
* **[OIP](https://github.com/oipwg/oip)**: The OIP daemon processes all Blocks and Transactions that exist in the Flo Blockchain, extracting OIP Records that were stored in Transactions. OIP has an API exposed on port `1606` [(available API endpoints)](https://github.com/oipwg/oip/blob/master/api.md).
* **[FLOd](https://github.com/bitspill/flod)**: A Go implmentation of a FLO Full node that OIP daemon connects to as its source of information. The RPC ports are not exposed.
* **[ElasticSearch](https://www.elastic.co/products/elasticsearch)**: ElasticSearch is used as the database backend for OIP daemon to allow for complex queries to be performed near instantaneous. ElasticSearch has its API exposed on port `9200`.
* **[Kibana](https://www.elastic.co/products/kibana)**: Kibana is installed to provide a convienent UI to view the OIP daemon database. Kibana has its API exposed on port `5601`.

## Available Versions

You can see all images available to pull from Docker Hub via the [Tags page](https://hub.docker.com/r/mediciland/oip/tags/).

## Usage Example
```
docker volume create oip

docker run -d \
  --mount source=oip,target=/data \
  -p 1606:1606 -p 5601:5601 -p 9200:9200 \
  --env HTTP_USER=oip --env HTTP_PASSWORD=mypassword \
  --env NETWORK=testnet --env ADDNODE=35.230.92.250 \
  --name=oip \
  mediciland/oip

docker logs --tail 5 -f oip
```

## Environment Variables

OIP uses Environment Variables to allow for configuration settings. You set the Env Variables in your `docker run` startup command. Here are the config settings offered by this image.

* **`HTTP_USER`**: [`String`] The username you wish to use for HTTP authentication for Kibana and Elasticsearch (Default `oipd`).
* **`HTTP_PASSWORD`**: [`String`] The password you wish to use for HTTP authentication for Kibana and Elasticsearch (Required).
* **`NETWORK`**: [`mainnet`|`testnet`|`regtest`] The Flo network you wish to run OIP on (Default `mainnet`).
* **`ADDNODE`**: [`ip-address`] An IP address of a Flo node to be used as a source of blocks. This is useful if you are running isolated networks, or if you are having a hard time connecting to the network.
* **`RPC_USER`**: [`String`] The RPC username for the Flod full node running inside the container.
* **`RPC_PASSWORD`**: [`String`] The RPC password for the Flod full node running inside the container.
* **`CUSTOM_BLACKLIST_FILTER`**: [`String` with format `label: remote url`] Add a custom blacklist filter url to the OIP config. Example `myfilter: http://myurl.com/blacklist.txt`.
* **`ELASTIC_RAM_SIZE`**: [`String` with format in MB or GB i.e. `3g`, `1500m`] The amount of RAM that should be alloted to ElasticSearch. Defaults to 1/4th of the available system ram. Minimum RAM: `1g`.

## Build Instructions
Want to build OIPd from it's source? We have created a simple docker build script that is able to build the OIPd binary and Docker Image very quickly, fully ensuring you have a non-tampered with copy!

### Build OIPd Binaries
First, you need to build the binaries for OIP daemon. You can do this by running the following script: `./ci/buildBinaries.sh`

### Build OIP Docker Image
Next, after the binaries have been built, build the docker image using `./ci/buildImage.sh`

## Hardware Requirements
In order to run the OIP docker image, it is suggested you meet the following minimum requirements for each of the following network types.

**mainnet**
* 2 Core CPU
* 7.5 GB RAM
* 50 GB Disk Space

**testnet**
* 1 Core CPU
* 3.75 GB RAM
* 50 GB Disk Space

**regtest**
* 1 Core CPU
* 3.75 GB RAM
* 5 GB Disk Space

## Development
To easily run a development server, ensure you have docker installed, and then run the script `start-dev.sh`. This will automatically build the binaries and docker image from scratch, and then run the image in a new docker container. It will then show you the logs. If you make a change to the source files, re-run the `start-dev.sh` script and it will automatically build the new version and start it up!

# Contacts
- bitspill, bitspill@oip.dev
- ohryan ryanachacon@gmail.com
- Chris Chrysostom, cchrysostom@mediciland.com
- Sky Young, skyoung@mediciland.com