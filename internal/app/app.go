package app

import (
	"context"

	"go-order-api/internal/infra/broker/rabbitmq"
	infraredis "go-order-api/internal/infra/cache/redis"
	"go-order-api/internal/infra/config"
	"go-order-api/internal/infra/http"
	"go-order-api/internal/infra/http/handler"
	"go-order-api/internal/infra/logger"
	"go-order-api/internal/infra/persistence/cached"
	"go-order-api/internal/infra/persistence/mongodb"
	usecase "go-order-api/internal/usecase/order"

	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	mongoClient   *mongo.Client
	redisClient   *goredis.Client
	messageBroker *rabbitmq.MessageBroker
	httpServer    *http.Server
	logger        *logger.ZapLogger
}

func New() (*App, error) {
	zapLogger := logger.NewZapLogger()

	config.LoadConfig(zapLogger)

	ctx := context.Background()

	mongoClient, err := mongodb.NewConnection(ctx, config.AppConfig.MongoURI, zapLogger)
	if err != nil {
		return nil, err
	}
	db := mongoClient.Database(config.AppConfig.DatabaseName)
	orderRepo := mongodb.NewOrderRepository(db, zapLogger)

	redisClient, err := infraredis.NewConnection(ctx, config.AppConfig.RedisAddr, config.AppConfig.RedisPassword, config.AppConfig.RedisDB, zapLogger)
	cache := infraredis.NewCache(ctx, redisClient, zapLogger)
	if err != nil {
		zapLogger.Warning("Redis unavailable at startup, continuing with cache client for automatic recovery")
	}
	cachedOrderRepo := cached.NewOrderRepository(orderRepo, cache, config.AppConfig.OrderCacheTTL, zapLogger)

	messageBroker, err := rabbitmq.NewMessageBroker(config.AppConfig.RabbitMQURL, zapLogger, rabbitmq.GetDefaultQueues())
	if err != nil {
		if redisClient != nil {
			_ = redisClient.Close()
		}
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}

	createOrderUC := usecase.NewCreateOrderUseCase(cachedOrderRepo)
	getOrderUC := usecase.NewGetOrderUseCase(cachedOrderRepo)
	updateOrderStatusUC := usecase.NewUpdateOrderStatusUseCase(cachedOrderRepo, messageBroker)

	orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, updateOrderStatusUC, zapLogger)

	mongoHealth := mongodb.NewMongoHealthChecker(mongoClient)
	rabbitHealth := rabbitmq.NewRabbitMQHealthChecker(messageBroker)
	healthHandler := handler.NewHealthHandler(zapLogger, mongoHealth, rabbitHealth)

	httpServer := http.NewServer(zapLogger, orderHandler, healthHandler)

	return &App{
		mongoClient:   mongoClient,
		redisClient:   redisClient,
		messageBroker: messageBroker,
		httpServer:    httpServer,
		logger:        zapLogger,
	}, nil
}

func (a *App) Run() {
	a.httpServer.Start()
}

func (a *App) Close() {
	if a.mongoClient != nil {
		_ = a.mongoClient.Disconnect(context.Background())
	}
	if a.redisClient != nil {
		_ = a.redisClient.Close()
	}
	if a.messageBroker != nil {
		a.messageBroker.Close()
	}
	if a.logger != nil {
		a.logger.Info("App resources closed")
		a.logger.Close()
	}
}
