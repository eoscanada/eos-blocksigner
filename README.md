EOS out-of-band block signer server
-----------------------------------

This is based on the `eosc vault` and works in combination with the
following `config.ini` from `nodeos`:

```
signature-provider=EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV=KEOSD:http://localhost:6666/v1/wallet/sign_digest
keosd-provider-timeout=5   # default value is 5 ms
```
