package app

import (
	"context"

	"go-order-api/internal/infra/broker/rabbitmq"
	"go-order-api/internal/infra/config"
	"go-order-api/internal/infra/http"
	"go-order-api/internal/infra/http/handler"
	"go-order-api/internal/infra/logger"
	"go-order-api/internal/infra/persistence/mongodb"
	usecase "go-order-api/internal/usecase/order"

	"go.mongodb.org/mongo-driver/mongo"
)

type App struct {
	mongoClient   *mongo.Client
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

	messageBroker, err := rabbitmq.NewMessageBroker(config.AppConfig.RabbitMQURL, zapLogger, rabbitmq.GetDefaultQueues())
	if err != nil {
		return nil, err
	}

	createOrderUC := usecase.NewCreateOrderUseCase(orderRepo)
	getOrderUC := usecase.NewGetOrderUseCase(orderRepo)
	updateOrderStatusUC := usecase.NewUpdateOrderStatusUseCase(orderRepo, messageBroker)

	orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, updateOrderStatusUC, zapLogger)

	mongoHealth := mongodb.NewMongoHealthChecker(mongoClient)
	rabbitHealth := rabbitmq.NewRabbitMQHealthChecker(messageBroker)
	healthHandler := handler.NewHealthHandler(zapLogger, mongoHealth, rabbitHealth)

	httpServer := http.NewServer(zapLogger, orderHandler, healthHandler)

	return &App{
		mongoClient:   mongoClient,
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
	if a.messageBroker != nil {
		a.messageBroker.Close()
	}
	if a.logger != nil {
		a.logger.Info("App resources closed")
		a.logger.Close()
	}
}
