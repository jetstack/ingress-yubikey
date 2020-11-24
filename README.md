## ingress-yubikey

This is a proof-of-concept **highly experimental!**
[Kubernetes Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/)
that terminates TLS using a certificate and Key from the PIV smartcard applet
on a Yubikey. This addresses a common complaint that Kubernetes Ingress
controllers have cluster-wide access to Secrets in order to retrieve 
TLS private Keys. With a hardware-backed key, the private key never exists 
in application memory

### Usage

Check you have a working yubikey before deploying:

```shell
./ingress-yubikey validate
```

`ingress-yubikey` watches for `networking.k8s.io/v1` Ingress objects
with their Ingress Class set to `ingress-yubikey` As the only goal is
to terminate TLS, path rules are ignored, but TLS hosts are matched
by parsing SNI.

For now, ingress-yubikey always uses the Digital Signature certificate
in slot 9c. Load this slot with a certificate, preferably with a private
Key generated on the Yubikey. Insert a Yubikey and run the ingress
controller using the manifest in `./deploy` as a guide. Volume mount
the smartcard appropriately.

#### PIN - protected keys

Again for now, the PIN for accessing the signing key can be provided with
the flag `--smartcard-pin` or environment variable
`INGRESS_YUBIKEY_SMARTCARD_PIN`.
