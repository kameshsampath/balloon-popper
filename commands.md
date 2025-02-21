# Some useful commands


## Verify RSA Keys

Assuming the passphrase is in `$PROJECT_HOME/keys/.pass`

```shell
openssl rsa -in keys/jwt-private.key -inform PEM -passin file:keys/.pass
```

Without key info

```shell
openssl rsa -in keys/jwt-private.key -inform PEM -passin file:keys/.pass -text -noout
```