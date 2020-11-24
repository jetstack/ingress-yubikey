## ingress-yubikey

This is a proof-of-concept **highly experimental!**
[Kubernetes Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/)
that terminates TLS using a certificate and key from the PIV smartcard applet
on a YubiKey. This addresses a common complaint that Kubernetes Ingress
controllers have cluster-wide access to secrets in order to retrieve 
TLS private keys. With a hardware-backed key, the private key never exists 
in application memory.

### Usage

Check you have a working yubikey before deploying:

```shell
./ingress-yubikey validate
```

If not, you can set up the PIV applet like so:

```shell
ykman piv reset
ykman piv generate-key -m 010203040506070801020304050607080102030405060708 -P 123456 -a ECCP256 --pin-policy NEVER --touch-policy NEVER 9c 9c.pub
ykman piv generate-csr -s your-hostname.com 9c 9c.pub 9c.csr
# Sign the CSR, even with a publicly trusted CA!
ykman piv import-certificate 9c 9c.pem
```

`ingress-yubikey` watches for `networking.k8s.io/v1` Ingress objects
with their Ingress Class set to `ingress-yubikey` As the only goal is
to terminate TLS, path rules are ignored, but TLS hosts are matched
by parsing SNI.

For now, ingress-yubikey always uses the Digital Signature certificate
in slot 9c. Insert an appropriately prepared YubiKey and run the ingress
controller using the manifest in `./deploy` as a guide. Volume mount
the smartcard device appropriately.

#### PIN - protected keys

Again for now, the PIN for accessing the signing key can be provided with
the flag `--smartcard-pin` or environment variable
`INGRESS_YUBIKEY_SMARTCARD_PIN`.
