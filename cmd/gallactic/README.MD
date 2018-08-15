gtxkey
======

gtxkey is a simple command-line tool working with gallactic keyfiles

# usage

### `gallactic gtxkey generate`

Generate new keyfile. This keyfile is encrypted with the password you provide while creating a new account.
You can specify the address type as 'ac' for accountAddress and 'va' for validatorAddress (By default account address [type: 'ac'] will be created)

Example:
```
gallactic gtxkey generate
gallactic gtxkey generate -t=va
```

### `gallactic gtxkey inspect <keyfile>`

Print various information about the keyfile.
Private key information can be printed using the `--private` or `-p` flag;
make sure you use this feature with great caution

Example:
```
gtxkey inspect acTmhDf7AjbCcvhkUAH3HFzV2aJw1rdJPwv.json
gtxkey inspect acTmhDf7AjbCcvhkUAH3HFzV2aJw1rdJPwv.json -p
```
*Note: By Default all your gallactic keys
will be stored in /tmp/gtxkeys directory in Ubuntu.

### `gallactic gtxkey sign <messagefile/message> <privatekey/keyfile-path>`

Signs the message with the private key. User can directly specify the private key to sign the message or specify the keyfilepath and password to decrypt the private key and sign the message.

Example:
```
gtxkey signmessage -m=hi -k=/tmp/gtxkeys/acVqy3VvA9RdWZ43ggCCRegq8XP3V47t8uX.json -a=<auth>
gtxkey signmessage -m=hi -k=/tmp/gtxkeys/acVqy3VvA9RdWZ43ggCCRegq8XP3V47t8uX.json
Passphrase:
gtxkey signmessage -f=/home/message/test1 -k=<keyfilepath>
Passphrase:
```

### `gallactic gtxkey verify <publicKey> <signature> <message>`

Verify the signature of the message.

Example:
```
gtxkey verify eb6c89611b4fd043fe87877b8c61ce740495e5d950be58ca5cf27dd99a89280b 9ab5a81c24e272e2009718b8c13aeaca32ffcf74be31b573b90e5f28885a52543fb3ee3c5028f94b919059de6292572631239a6471d4d70db3ed69c7390ac90a -m=hi
Signature Verification successfull!

gtxkey verify eb6c89611b4fd043fe87877b8c61ce740495e5d950be58ca5cf27dd99a89280b 9ab5a81c24e272e2009718b8c13aeaca32ffcf74be31b573b90e5f28885a52543fb3ee3c5028f94b919059de6292572631239a6471d4d70db3ed69c7390ac90a -f=/home/message/test1
Signature Verification successfull!
```

### `gallactic gtxkey changepassword`

Change the password of the keyfile which is used for encrytption and decrytption.