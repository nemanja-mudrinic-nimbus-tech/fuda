package login

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
)

type sealer struct{ aead cipher.AEAD }

func newSealer(secret string) (sealer, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return sealer{}, err
	}
	aead, err := cipher.NewGCM(block)
	return sealer{aead}, err
}

func (s sealer) seal(v any) (string, error) {
	plain, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(s.aead.Seal(nonce, nonce, plain, nil)), nil
}

func (s sealer) open(text string, v any) error {
	raw, err := base64.RawURLEncoding.DecodeString(text)
	if err != nil {
		return err
	}
	if len(raw) < s.aead.NonceSize() {
		return errors.New("sealed value is too short")
	}
	nonce, sealed := raw[:s.aead.NonceSize()], raw[s.aead.NonceSize():]
	plain, err := s.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(plain, v)
}
