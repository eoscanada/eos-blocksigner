EOS out-of-band block signer server
-----------------------------------

The EOS.IO Software's `nodeos` program signs blocks with keys it holds
in memory.

Introduced slightly before th 1.0 release, is an Out-of-band signing
method.  It involves setting up your configuration with something like
this:

```
signature-provider=EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV=KEOSD:http://localhost:6666/v1/wallet/sign_digest
keosd-provider-timeout=5   # default value is 5 ms
```

This means that `nodeos` can sign blocks with the private key
corresponding to `EOS6MR...W5CV` through a `keosd`-compatible program.

`eos-blocksigner` is such a program, and integrates with `eosc`, the
wallet and command-line tool.

WARNING: you do *NOT* want to expose that software to any public
endpoint, _even in an internal network_. It should run and listen
strictly on a loopback interface. The `sign_digest` endpoint can
actually sign *anything* with the associated private key. If you are
using the same private key for `owner` and/or `active` permissions on
some accounts (which you should *not*), then any transaction can be
signed with the `/v1/wallet/sign_digest` endpoint.

## Two modes of operation

As of May 2018, `eos-blocksigner` has two modes of operation:

1. using a vault encrypted through Google Cloud Platform's Key Management System
2. using a plain-text private keys file

As demand grows, we can add more strategies, like AWS's KMS system,
passphrase-encrypted vaults, or some other HSM systems.


## GCP KMS integration

To use the KMS-GCP strategy, create a vault locally using `eosc` this way:

```
$ eosc vault create --import \
                    --vault-type kms-gcp \
                    --comment "Block signing key vault" \
                    --kms-gcp-keypath projects/PROJNAME/locations/LOC/keyRings/RINGNAME/cryptoKeys/KEYNAME
...
```

This implies you have authenticated through `gcloud` and have
permissions to _Encrypt_ using KMS, in the specified project and
keyring.

You can then drop the `eosc-vault.json` wallet on your production
infrastructure, and run `eos-blocksigner` with these parameters:

```
$ eos-blocksigner --wallet-path path/to/eosc-vault.json \
                  --kms-gcp-keypath projects/PROJNAME/locations/LOC/keyRings/RINGNAME/cryptoKeys/KEYNAME
Listening on 127.0.0.1:6666
```


## Plain-text private keys file

This is a method which is not very secure, yet is still more secure
than keeping your private keys in plain-text in your `config.ini`.

Remote code execution vulnerabilities often allow reading the process'
memory easily. Having the out-of-band (read: separate process) signing
server already makes it more complex to access memory with your
private keys.  With proper isolation (containers, network access, and
`eos-blocksigner`), you can mitigate the risk of leaking your private
keys through an unforeseen `nodeos` vulnerability.

The `--keys-file` is a simple file that looks like this (`myfile.keys`):

```
5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFFF
5KQwrPbwdL6PhXujxW37FSSQZ1JiwsST4cqQzDeyXtP79zkvFD3 // This matches EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV
```

It has one private key per line. Anything after an optional whitespace
is ignored.

With a keys-file, you don't need an `eosc-vault.json`, and can run:

```
$ eos-blocksigner --keys-file=myfile.keys
Listening on 127.0.0.1:6666
```

# License

MIT
