package adapter

import (
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/zeals-co-ltd/shopify-app-example/internal/config"
	"github.com/zeals-co-ltd/shopify-app-example/internal/usecase"
	"github.com/zeals-co-ltd/shopify-app-example/pkg/shopify"
)

const (
	scopes = "read_products,write_products"
)

type HttpServer interface {
	Run(string) error
}

type httpServer struct {
	shopifyClient shopify.Client
	scopes        string
	apiKey        string
	apiSecret     string
	serverUrl     string
	usecase       usecase.ShopifyUsecase
}

func NewHttpServer(
	shopifyClient shopify.Client,
	shopifyUsecase usecase.ShopifyUsecase,
) (HttpServer, error) {
	apiSecret, err := config.MustGet("SHOPIFY_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}

	apiKey, err := config.MustGet("SHOPIFY_CLIENT_ID")
	if err != nil {
		log.Err(err).Msg("failed to get SHOPIFY_CLIENT_ID")
		return nil, err
	}

	serverUrl, err := config.MustGet("SERVER_URL")
	if err != nil {
		return nil, err
	}
	return &httpServer{
		shopifyClient: shopifyClient,
		scopes:        scopes,
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		serverUrl:     serverUrl,
		usecase:       shopifyUsecase,
	}, nil
}

func (h *httpServer) Run(port string) error {
	http.HandleFunc("/shopify", h.shopifyHandler())
	http.HandleFunc("/shopify/callback", h.shopifyCallbackHandler())
	http.HandleFunc("/app", h.appHandler())
	http.HandleFunc("/webhook", h.webhookHandler())

	return http.ListenAndServe(port, nil)
}

func (h *httpServer) shopifyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := usecase.RequestAuthorizationRequest{
			Url: r.URL,
		}

		redirectUrl, err := h.usecase.RequestAuthorization(r.Context(), req)
		if err != nil {
			response := ErrorResponse{Errors: err.Error()}
			w.Write(response.ToJson())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, redirectUrl, http.StatusSeeOther)
	}
}

func (h *httpServer) shopifyCallbackHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h.usecase.Authorize(r.Context(), usecase.AuthorizeRequest{Url: r.URL})
		if err != nil {
			response := ErrorResponse{Errors: err.Error()}
			w.Write(response.ToJson())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// TODO: pass session
		http.Redirect(w, r, "/app", http.StatusSeeOther)
	}
}

func (h *httpServer) appHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tmpl, err = template.ParseFiles("index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var data = map[string]interface{}{
			"title": "Shopify app testing",
			"name":  "Batman",
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *httpServer) webhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: call usecase

		w.WriteHeader(http.StatusOK)
	}
}
