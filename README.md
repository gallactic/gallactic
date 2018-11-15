# Gallactic [![Build Status](https://api.travis-ci.org/gallactic/gallactic.svg?branch=master)](https://travis-ci.org/gallactic/gallactic)

*Gallactic blockchain with [SputnikVM](https://github.com/gallactic/sputnikvm) and [Tendermint](https://github.com/tendermint/tendermint) consensus engine*

## Compiling the code

You need to make sure you have install [Go](https://golang.org/) (version 1.10.1 or higher) and [rust](https://www.rust-lang.org). After installing them, you can follow these steps to compile and build the gallactic project:

```bash
mkdir -p $GOPATH/src/github.com/gallactic/gallactic
cd $GOPATH/src/github.com/gallactic
git clone https://github.com/gallactic/gallactic.git .
make
```

Run `gallactic version` to make sure gallactic is properly compiled and installed in your machine.

## Running Gallactic

### Initialize

Initialize the working directory by running:

 ```bash
 gallactic init -w=<workspace_directory>
 ```

 This command will create config.toml, genesis.json and private key for validator.

### Run

For running a Gallactic node, use:

```bash
gallactic start -w=<workspace_directory>
```

This command will ask you to enter the private key of the validator. Enter the private key (priv_key) of the validator, as provided by the init command above.
The Gallactic blockchain starts immediately, upon successful acceptance of the private key.

## Usage of Docker

Install [Docker](https://www.docker.com/) and run the following commands to build the docker file:

```bash
cd $GOPATH/src/github.com/gallactic/gallactic
docker build . --tag gallactic
```

Then you can execute the Gallactic blockchain, using the docker:

```bash
# Initializing the working directory
docker run -it --rm -v "/tmp/chain1:/gallactic" gallactic init
# Starting the blockchain
docker run -it --rm -v "/tmp/chain1:/gallactic" -p 1337:1337 -p 50051:50051 -p 46656:46656 gallactic start
```

## Contribution

Thanks for considering to contribute in Gallactic project! You can read [here](https://github.com/gallactic/gallactic/wiki/Contribute) to read how you can contribute on Gallactic project.

## License

The Gallactic blockchain is under MIT license.