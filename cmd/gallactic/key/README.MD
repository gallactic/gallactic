# gallactic key

`gallactic key` is a simple command-line tool to work with gallactic key files.

## Usage

### `gallactic key generate`

Generate new key file. This key file is encrypted with the password you provide while creating a new account.
You can specify the address type as 'ac' for account address and 'va' for validator address (By default account address will be created)

Example:

```bash
gallactic key generate
gallactic key generate -t va
```

### `gallactic key inspect KEYFILE`

Print various information about the given key file.
Private key information can be printed using the `-e` flag. make sure you use this feature with great caution.

Example:

```bash
gallactic key inspect ~/gallactic/keystore/acLjwzaYPc8Nmbj5AKp2vMp3GQoGfHg1t3A.json
gallactic key inspect ~/gallactic/keystore/acLjwzaYPc8Nmbj5AKp2vMp3GQoGfHg1t3A.json -e
```

### `gallactic key sign <file/message> <privatekey/keyfile>`

Signs the message with the private key. User can directly specify the private key to sign the message or specify the a key file and password to decrypt the private key and sign the message.

Example:

```bash
gallactic key sign -m hello -k ~/gallactic/keystore/acLjwzaYPc8Nmbj5AKp2vMp3GQoGfHg1t3A.json
```

### `gallactic key verify PUBLICKEY SIGNATURE <file/message>`

Verify the signature of a message.

Example:

```bash
gallactic key verify pkwgaQWBM3oJ5mw2n4UT56teJqGX8jCie5UWiNKdHwXTPGiRWy3 94f32a20f649215d0f983ae1f6a6bde1cd49180bf24261385936e82c5fa54e830a489b80d718b1d470b19e35376b39286ccb72f8f8746fe7d728398951a8750a -m hello
```

### `gallactic key change-auth KEYFILE`

Change the password of the key file which is used for encryption and decryption.

Example:

```bash
gallactic key change-auth ~/gallactic/keystore/acLjwzaYPc8Nmbj5AKp2vMp3GQoGfHg1t3A.json
```