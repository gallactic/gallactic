# Gallactic [![Build Status](https://api.travis-ci.org/gallactic/gallactic.svg?branch=master)](https://travis-ci.org/gallactic/gallactic)
*Gallactic blockchain with Hyperledger Burrow EVM and Tendermint consensus engine*

## Compiling the code
You need to make sure you have install go version 1.10.1 or higher. Technically your go should have ed25519 signature malleability fix.

Follow these steps to compile and build the gallactic executable:

```
# Get the source code
mkdir -p $GOPATH/src/github.com/gallactic
cd $GOPATH/src/github.com/gallactic
git clone https://github.com/gallactic/gallactic.git
cd gallactic

make
make install
```

Run `gallactic version` to make sure gallactic is properly compiled and installed in yur machine.

NOTE:

prevent signature malleability

ed25519: require canonical signatures

https://go-review.googlesource.com/c/crypto/+/100436

## Contribution
Thanks for considering to contribute in Gallactic project!

## License
The Gallactic blockchain is under MIT <TODO: link> license.
