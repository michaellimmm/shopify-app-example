package shopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
)

func VerifyAuthUrl(u *url.URL, apiSecret string) (bool, error) {
	val := u.Query()
	messageMAC := val.Get("hmac")

	// remove hmac and signature
	val.Del("hmac")
	val.Del("signature")

	message, err := url.QueryUnescape(val.Encode())
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(message))
	expectedMac := mac.Sum(nil)

	actualMac, err := hex.DecodeString(messageMAC)
	if err != nil {
		return false, err
	}

	return hmac.Equal(expectedMac, actualMac), nil
}
