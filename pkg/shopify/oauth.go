package shopify

import (
	"encoding/json"
	"net/url"
)

type OauthService interface {
	GetAccessToken(shop string, code string) (*TokenResponse, error)
}

type TokenRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

func (r *TokenRequest) ToBytes() []byte {
	body, _ := json.Marshal(r)
	return body
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
}

func (c *client) GetAccessToken(shop string, code string) (*TokenResponse, error) {
	requestUrl, err := url.Parse("https://" + shop)
	if err != nil {
		return nil, err
	}

	requestUrl.Path = "/admin/oauth/access_token"

	data := &TokenRequest{
		ClientId:     c.apiKey,
		ClientSecret: c.apiSecret,
		Code:         code,
	}

	req, err := NewRequest("POST", requestUrl, "", data)
	if err != nil {
		return nil, err
	}

	token := new(TokenResponse)
	err = c.SendRequest(req, token)
	if err != nil {
		return nil, err
	}

	return token, nil
}
