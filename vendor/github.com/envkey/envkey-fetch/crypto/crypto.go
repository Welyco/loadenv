package crypto

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"reflect"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/clearsign"
)

func ReadPrivkey(encryptedPrivkeyArmored, pw []byte) (openpgp.EntityList, error) {
	// Read the private key
	entityList, err := ReadArmoredKey(encryptedPrivkeyArmored)
	if err != nil {
		return nil, err
	}
	entity := entityList[0]

	// Get the passphrase and read the private key.
	entity.PrivateKey.Decrypt(pw)
	for _, subkey := range entity.Subkeys {
		subkey.PrivateKey.Decrypt(pw)
	}

	return entityList, nil
}

func ReadArmoredKey(armoredKey []byte) (openpgp.EntityList, error) {
	keyringFileBuffer := bytes.NewBuffer(armoredKey)
	return openpgp.ReadArmoredKeyRing(keyringFileBuffer)
}

func MakeKeyring(decryptedPrivkey openpgp.EntityList, pubkeyArmored []byte) (openpgp.EntityList, error) {
	pubkey, err := ReadArmoredKey(pubkeyArmored)
	if err != nil {
		return nil, err
	}
	return append(decryptedPrivkey, pubkey...), nil
}

func VerifyPubkeyWithPrivkey(pubkey, decryptedPrivkey openpgp.EntityList) error {
	var encrypted, decrypted []byte
	var err error

	msg := []byte("test message")

	encrypted, err = Encrypt(msg, pubkey)
	if err != nil {
		return err
	}

	decrypted, err = Decrypt(encrypted, decryptedPrivkey)
	if err != nil {
		return err
	}

	if bytes.Equal(msg, decrypted) {
		return nil
	} else {
		return errors.New("Decrypted message does not match original message.")
	}
}

func Encrypt(msg []byte, pubkeys openpgp.EntityList) ([]byte, error) {
	var encCloser, armorCloser io.WriteCloser
	var err error

	encbuf := new(bytes.Buffer)
	encCloser, err = openpgp.Encrypt(encbuf, pubkeys, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	_, err = encCloser.Write(msg)
	if err != nil {
		return nil, err
	}

	err = encCloser.Close()
	if err != nil {
		return nil, err
	}

	armorbuf := new(bytes.Buffer)
	armorCloser, err = armor.Encode(armorbuf, "PGP MESSAGE", nil)
	if err != nil {
		return nil, err
	}

	_, err = armorCloser.Write(encbuf.Bytes())

	err = armorCloser.Close()
	if err != nil {
		return nil, err
	}

	return armorbuf.Bytes(), nil
}

func Decrypt(cipherArmored []byte, keys openpgp.EntityList) ([]byte, error) {
	if !(len(keys) == 1 && keys[0].PrivateKey != nil) {
		return nil, errors.New("Requires a single private key.")
	}
	return readMessage(cipherArmored, keys)
}

func DecryptAndVerify(cipherArmored []byte, keys openpgp.EntityList) ([]byte, error) {
	if !(len(keys) == 2 && keys[0].PrivateKey != nil && keys[1].PrivateKey == nil) {
		return nil, errors.New("Requires a single private key and a single public key.")
	}

	return readMessage(cipherArmored, keys)
}

func VerifySignedCleartext(message []byte, keys openpgp.EntityList) ([]byte, error) {
	if !(len(keys) == 1 && keys[0].PrivateKey == nil) {
		return nil, errors.New("Requires a single public key.")
	}

	block, _ := clearsign.Decode(message)
	_, err := openpgp.CheckDetachedSignature(keys, bytes.NewBuffer(block.Bytes), block.ArmoredSignature.Body)

	if err != nil {
		return nil, err
	}

	return block.Bytes, nil
}

func VerifyPubkeySignature(signedPubkey, signerPubkey openpgp.EntityList) error {
	signedKey := signedPubkey[0]
	signerKey := signerPubkey[0]

	signedIdentityName := reflect.ValueOf(signedKey.Identities).MapKeys()[0].Interface().(string)
	signature := signedKey.Identities[signedIdentityName].Signatures[0]

	return signerKey.PrimaryKey.VerifyUserIdSignature(signedIdentityName, signedKey.PrimaryKey, signature)
}

func VerifyPubkeyArmoredSignature(signedPubkeyArmored, signerPubkeyArmored []byte) error {
	signedPubkey, err := ReadArmoredKey(signedPubkeyArmored)
	if err != nil {
		return err
	}

	signerPubkey, err := ReadArmoredKey(signerPubkeyArmored)
	if err != nil {
		return err
	}

	return VerifyPubkeySignature(signedPubkey, signerPubkey)
}

func readMessage(armoredMessage []byte, keys openpgp.EntityList) ([]byte, error) {
	// Decode armored message
	decbuf := bytes.NewBuffer(armoredMessage)
	result, err := armor.Decode(decbuf)
	if err != nil {
		return nil, err
	}

	// Decrypt with private key
	md, err := openpgp.ReadMessage(result.Body, keys, nil, nil)
	if err != nil {
		return nil, err
	}

	// If pubkey included, verify
	if len(keys) == 2 {
		if md.SignedBy == nil || md.SignedBy.PublicKey == nil {
			return nil, errors.New("Verifying public key included, but message is not signed.")
		} else if md.SignedBy.PublicKey.Fingerprint != keys[1].PrimaryKey.Fingerprint {
			return nil, errors.New("Signature pubkey doesn't match signing pubkey.")
		}
	}

	bytes, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, err
	}
	if md.SignatureError != nil {
		return nil, md.SignatureError
	}

	return bytes, nil
}
