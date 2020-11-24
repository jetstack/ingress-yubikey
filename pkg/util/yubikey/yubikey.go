package yubikey

import (
	"github.com/go-piv/piv-go/piv"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func Validate() (*piv.YubiKey, error){
	smartCards, err := piv.Cards()
	if err != nil {
		return nil, err
	}
	var yk *piv.YubiKey
	for _, card := range smartCards {
		yk, err = piv.Open(card)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't open yubikey")
		} else {
			break
		}
	}
	if yk == nil {
		return nil, errors.New("no usable smartcard found")
	}
	cert, err := yk.Certificate(piv.SlotSignature)
	if err != nil {
		yk.Close()
		return nil, errors.Wrap(err, "no cert in Signature slot")
	}
	_, err = yk.PrivateKey(piv.SlotSignature, cert.PublicKey, piv.KeyAuth{PIN: viper.GetString("smartcard-pin")})
	if err != nil {
		yk.Close()
		return nil, errors.Wrap(err, "couldn't unlock private key")
	}
	return yk, nil
}
