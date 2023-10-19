package shopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/zeals-co-ltd/shopify-app-example/internal/config"
)

const (
	apiVersion = "2023-07"
)

type Client interface {
	WebhookService
	OauthService
}

type client struct {
	httpClient *http.Client
	apiKey     string
	apiSecret  string
}

func NewClient(httpClient *http.Client) (Client, error) {
	apiSecret, err := config.MustGet("SHOPIFY_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}

	apiKey, err := config.MustGet("SHOPIFY_CLIENT_ID")
	if err != nil {
		return nil, err
	}

	return &client{
		httpClient: httpClient,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
	}, nil
}

func NewRequest(
	method string,
	url *url.URL,
	accessToken string,
	body interface{},
) (*http.Request, error) {
	var payload io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		payload = bytes.NewBuffer(buf)
	}

	req, err := http.NewRequest(method, url.String(), payload)
	if err != nil {
		log.Error().Err(err).Msg("error")
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	if accessToken != "" {
		req.Header.Add("X-Shopify-Access-Token", accessToken)
	}

	return req, nil
}

func (c *client) SendRequest(request *http.Request, response interface{}) error {
	res, err := c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
		return fmt.Errorf("error with status code %d", res.StatusCode)
	}

	if response != nil {
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(response)
		if err != nil {
			return err
		}
	}

	return nil
}
