# Gallactic [![Build Status](https://api.travis-ci.org/gallactic/gallactic.svg?branch=master)](https://travis-ci.org/gallactic/gallactic) 
*Gallactic blockchain with Hyperledger Burrow EVM and Tendermint consensus engine*

### Compiling the code
------
You need to make sure you have install go version 1.10.1 or higher. Technically your go should have ed25519 signature malleability fix.

 * Follow these steps to compile and build the gallactic executable:

```
Get the source code
mkdir -p $GOPATH/src/github.com/gallactic
cd $GOPATH/src/github.com/gallactic
git clone https://github.com/gallactic/gallactic.git
cd gallactic

make
make install
```
### Running the Gallactic Block Chain
--------
 **gallactic init -w=<workspace_folder_path>**

 This command will:
1. Create 3 files in the folder, path of which is given above, namely config.toml, genesis.json.
2. Display the private key (let’s call it priv_key) and validator address 

 **gallactic start -w=<workspace_folder_path>**

> This command will prompt to enter the private key of the validator. Enter the private key (priv_key) of the validator, as provided by the init command above.
The Gallactic blockchain starts immediately, upon successful acceptance of the private key.


### Usage of Docker
------
    Get the source code*
	mkdir -p $GOPATH/src/github.com/gallactic
	cd $GOPATH/src/github.com/gallactic
	git clone https://github.com/gallactic/gallactic.git
	cd gallactic

> [Install](https://www.docker.com/) the docker, as per the OS version of the system.

 * Build the docker file to run the Gallactic blockchain, as follows: 
_____

     docker build $GOPATH/src/github.com/gallactic/gallactic/ --tag gallactic

The argument **tag** creates a repository titled **GALLACTIC** inside the docker.

* Run the following command to display the Gallactic repository, which is present within the docker:

> docker images

 * Run the following commands to execute the Gallactic blockchain, using the docker:
For Mac OSX users:

     docker run -it --rm -v "/tmp/chain1:/gallactic"  gallactic init -w=/gallactic

> For users other than Mac OSX:

docker run gallactic init -w=<workspace_folder_path>

This command will:
1. Create 3 files in the folder, path of which is given above, namely config.toml, genesis.json, and pv_validator.json
2. Display the private key (let’s call it priv_key) and validator address 

     **gallactic start -w=<workspace_folder_path>**

> This command will prompt to enter the private key of the validator. Enter the private key (priv_key) of the validator, as provided by the init command above.
The Gallactic blockchain starts immediately, upon successful acceptance of the private key.

> Run `gallactic version` to make sure gallactic is properly compiled and installed in your machine.

NOTE:

 > prevent signature malleability

 > ed25519: require canonical signatures

 >  https://go-review.googlesource.com/c/crypto/+/100436

## Contribution
*Thanks for considering to contribute in Gallactic project!*

## License
*The Gallactic blockchain is under MIT <TODO: link> license.*
