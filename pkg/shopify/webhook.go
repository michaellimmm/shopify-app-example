package shopify

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

const webhooksBasePath = "webhooks"

type WebhookService interface {
	ListWebhook(shop string, accessToken string, options interface{}) ([]Webhook, error)
	GetWebhook(shop string, id int64, accessToken string, options interface{}) (*Webhook, error)
	CreateWebhook(shop string, accessToken string, webhook Webhook) (*Webhook, error)
	DeleteWebhook(shop string, accessToken string, id int64) error
}

type Webhook struct {
	ID                         int64      `json:"id"`
	Address                    string     `json:"address"`
	Topic                      string     `json:"topic"`
	Format                     string     `json:"format"`
	CreatedAt                  *time.Time `json:"created_at,omitempty"`
	UpdatedAt                  *time.Time `json:"updated_at,omitempty"`
	Fields                     []string   `json:"fields"`
	MetafieldNamespaces        []string   `json:"metafield_namespaces"`
	PrivateMetafieldNamespaces []string   `json:"private_metafield_namespaces"`
	ApiVersion                 string     `json:"api_version,omitempty"`
}

type WebhookResource struct {
	Webhook *Webhook `json:"webhook"`
}

func (r *WebhookResource) ToBytes() []byte {
	result, _ := json.Marshal(&r)
	return result
}

type WebhookResources struct {
	Webhooks []Webhook `json:"webhooks"`
}

func (r *WebhookResources) ToBytes() []byte {
	result, _ := json.Marshal(r)
	return result
}

// WebhookOptions can be used for filtering webhooks on a List request.
type WebhookOptions struct {
	Address string `url:"address,omitempty"`
	Topic   string `url:"topic,omitempty"`
}

func (c *client) ListWebhook(shop, accessToken string, options interface{}) ([]Webhook, error) {
	requestUrl, err := url.Parse("https://" + shop)
	if err != nil {
		return nil, err
	}

	requestUrl.Path = fmt.Sprintf("admin/api/%s/%s.json", apiVersion, webhooksBasePath)

	req, err := NewRequest("GET", requestUrl, accessToken, nil)
	if err != nil {
		return nil, err
	}

	result := new(WebhookResources)
	err = c.SendRequest(req, result)
	if err != nil {
		return nil, err
	}

	return result.Webhooks, nil
}

func (c *client) GetWebhook(
	shop string,
	id int64,
	accessToken string,
	options interface{},
) (*Webhook, error) {
	requestUrl, err := url.Parse("https://" + shop)
	if err != nil {
		return nil, err
	}
	requestUrl.Path = fmt.Sprintf("admin/api/%s/%s/%d.json", apiVersion, webhooksBasePath, id)

	req, err := NewRequest("GET", requestUrl, accessToken, nil)
	if err != nil {
		return nil, err
	}

	result := new(WebhookResource)
	err = c.SendRequest(req, result)
	if err != nil {
		return nil, err
	}

	return result.Webhook, nil
}

func (c *client) CreateWebhook(
	shop string,
	accessToken string,
	webhook Webhook,
) (*Webhook, error) {
	requestUrl, err := url.Parse("https://" + shop)
	if err != nil {
		return nil, err
	}
	requestUrl.Path = fmt.Sprintf("admin/api/%s/%s.json", apiVersion, webhooksBasePath)

	request := WebhookResource{Webhook: &webhook}
	req, err := NewRequest("POST", requestUrl, accessToken, request)
	if err != nil {
		return nil, err
	}

	result := new(WebhookResource)
	err = c.SendRequest(req, result)
	if err != nil {
		return nil, err
	}

	return result.Webhook, nil
}

func (c *client) DeleteWebhook(
	shop string,
	accessToken string,
	id int64,
) error {
	requestUrl, err := url.Parse("https://" + shop)
	if err != nil {
		return err
	}
	requestUrl.Path = fmt.Sprintf("admin/api/%s/%s/%d.json", apiVersion, webhooksBasePath, id)

	log.Info().Str("url", requestUrl.String()).Msg("delete url")

	req, err := NewRequest("DELETE", requestUrl, accessToken, nil)
	if err != nil {
		return err
	}

	err = c.SendRequest(req, nil)
	if err != nil {
		return err
	}

	return nil
}
