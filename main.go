package main

import (
	"context"
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zeals-co-ltd/shopify-app-example/internal/adapter"
	"github.com/zeals-co-ltd/shopify-app-example/internal/config"
	"github.com/zeals-co-ltd/shopify-app-example/internal/repository"
	"github.com/zeals-co-ltd/shopify-app-example/internal/usecase"
	"github.com/zeals-co-ltd/shopify-app-example/pkg/shopify"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	config.Load(".env")

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	defer func() {
		if r := recover(); r != nil {
			log.Error().Any("error", r).Any("stack", debug.Stack()).Msg("panic")
		}
	}()

	httpClient := &http.Client{}
	shopifyClient, err := shopify.NewClient(httpClient)
	if err != nil {
		log.Err(err).Msg("failed to initiate ShopifyClient")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Err(err).Msg("failed to initiate mongo client")
		return
	}
	defer mongoClient.Disconnect(ctx)

	// repository
	authRepository, err := repository.NewAuthRepository(mongoClient.Database("shopify_db"))
	if err != nil {
		log.Err(err).Msg("failed to initiate repository")
		return
	}

	// usecase
	shopifyUsecase, err := usecase.NewShopifyUsecase(shopifyClient, authRepository)
	if err != nil {
		log.Err(err).Msg("failed to initiate shopifyUsecase")
		return
	}

	httpServer, err := adapter.NewHttpServer(shopifyClient, shopifyUsecase)
	if err != nil {
		log.Err(err).Msg("failed to initiate HttpServer")
		return
	}

	err = httpServer.Run(":3434")
	if err != nil {
		log.Err(err).Msg("failed run server")
		return
	}
}
