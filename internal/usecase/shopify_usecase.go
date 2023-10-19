package usecase

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/zeals-co-ltd/shopify-app-example/internal/config"
	"github.com/zeals-co-ltd/shopify-app-example/internal/model"
	"github.com/zeals-co-ltd/shopify-app-example/internal/repository"
	"github.com/zeals-co-ltd/shopify-app-example/pkg/shopify"
)

const (
	scopes = "read_products,write_products"
)

type webhookTopic string

const (
	productCreatedTopic webhookTopic = "products/create"
	productUpdatedTopic webhookTopic = "products/update"
	productDeletedTopic webhookTopic = "products/delete"
	appUninstalledTopic webhookTopic = "app/uninstalled"
)

type ShopifyUsecase interface {
	RequestAuthorization(ctx context.Context, req RequestAuthorizationRequest) (string, error)
	Authorize(ctx context.Context, req AuthorizeRequest) error
}

type shopifyUsecase struct {
	shopifyClient  shopify.Client
	authRepository repository.AuthRepository
	apiKey         string
	apiSecret      string
	serverUrl      string
}

func NewShopifyUsecase(
	shopifyClient shopify.Client,
	authRepository repository.AuthRepository,
) (ShopifyUsecase, error) {
	apiSecret, err := config.MustGet("SHOPIFY_CLIENT_SECRET")
	if err != nil {
		return nil, errors.New("failed to get SHOPIFY_CLIENT_SECRET")
	}

	apiKey, err := config.MustGet("SHOPIFY_CLIENT_ID")
	if err != nil {
		return nil, errors.New("failed to get SHOPIFY_CLIENT_ID")
	}

	serverUrl, err := config.MustGet("SERVER_URL")
	if err != nil {
		return nil, errors.New("failed to get SERVER_URL")
	}

	return &shopifyUsecase{
		shopifyClient:  shopifyClient,
		authRepository: authRepository,
		apiSecret:      apiSecret,
		apiKey:         apiKey,
		serverUrl:      serverUrl,
	}, nil
}

type RequestAuthorizationRequest struct {
	Url *url.URL
}

func (r *RequestAuthorizationRequest) GetShop() string {
	val := r.Url.Query()
	return val.Get("shop")
}

func (r *RequestAuthorizationRequest) Validate() error {
	val := r.Url.Query()
	shop := val.Get("shop")
	if shop == "" {
		return errors.New(`missing "shop" parameter`)
	}

	return nil
}

func (uc *shopifyUsecase) RequestAuthorization(ctx context.Context, req RequestAuthorizationRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", err
	}

	nonce := uuid.NewString()
	redirectedUrl := uc.serverUrl + "/shopify/callback"
	shopUrl, err := url.Parse("https://" + req.GetShop())
	if err != nil {
		return "", err
	}
	shopUrl.Path = "/admin/oauth/authorize"
	query := shopUrl.Query()
	query.Set("client_id", uc.apiKey)
	query.Set("scope", scopes)
	query.Set("state", nonce)
	query.Set("redirect_uri", redirectedUrl)
	shopUrl.RawQuery = query.Encode()

	return shopUrl.String(), nil
}

type AuthorizeRequest struct {
	Url *url.URL
}

func (r *AuthorizeRequest) GetShop() string {
	val := r.Url.Query()
	return val.Get("shop")
}

func (r *AuthorizeRequest) GetCode() string {
	val := r.Url.Query()
	return val.Get("code")
}

func (r *AuthorizeRequest) Validate(apiSecret string) error {
	val := r.Url.Query()
	shop := val.Get("shop")
	if shop == "" {
		return errors.New(`missing "shop" parameter`)
	}

	code := val.Get("code")
	if code == "" {
		return errors.New(`missing "code" parameter`)
	}

	if ok, err := shopify.VerifyAuthUrl(r.Url, apiSecret); !ok || err != nil {
		return errors.New("hmac is not match")
	}

	return nil
}

func (uc *shopifyUsecase) Authorize(ctx context.Context, req AuthorizeRequest) error {
	if err := req.Validate(uc.apiSecret); err != nil {
		return err
	}

	auth, err := uc.authRepository.FindByShop(ctx, req.GetShop())
	if err != nil {
		return err
	}

	if !auth.IsEmpty() {
		webhooks, err := uc.shopifyClient.ListWebhook(auth.Shop, auth.AccessToken, nil)
		if err != nil {
			log.Error().Err(err).Msg("failed to get ListWebhook")
			return err
		}
		log.Info().Any("webhooks", webhooks).Msg("webhook message")

		return nil
	}

	token, err := uc.shopifyClient.GetAccessToken(req.GetShop(), req.GetCode())
	if err != nil {
		log.Error().Err(err).Msg("failed get access token")
		return err
	}

	_, err = uc.authRepository.Save(ctx, model.ShopifyAuth{
		Shop:        req.GetShop(),
		AccessToken: token.AccessToken,
	})
	if err != nil {
		return err
	}

	uc.registerWebhook(req.GetShop(), token.AccessToken)

	return nil
}

func (uc *shopifyUsecase) registerWebhook(shop, accessToken string) {
	topics := []webhookTopic{
		productCreatedTopic,
		productUpdatedTopic,
		productDeletedTopic,
		appUninstalledTopic,
	}

	var wg sync.WaitGroup

	for _, topic := range topics {
		wg.Add(1)

		go func(topic string) {
			defer wg.Done()

			request := shopify.Webhook{
				Address: uc.serverUrl + "/webhook",
				Topic:   string(topic),
				Format:  "json",
			}
			webhook, err := uc.shopifyClient.CreateWebhook(shop, accessToken, request)
			if err != nil {
				log.Err(err).Msg(webhook.Topic + "error create product/create webhook")
			}
			log.Info().Any("webhook", webhook).Msg(webhook.Topic + " webhook created")
		}(string(topic))
	}

	// uninstall webhook
	// for _, webhook := range webhooks {
	// 	wg.Add(1)

	// 	go func(id int64) {
	// 		defer wg.Done()

	// 		err := uc.shopifyClient.DeleteWebhook(shop, accessToken, id)
	// 		if err != nil {
	// 			log.Error().Err(err).Msg("error delete webhook")
	// 		}
	// 	}(webhook.ID)
	// }

	wg.Wait()
}
