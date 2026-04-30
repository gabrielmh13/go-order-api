# Contexto do Projeto

Este documento descreve o funcionamento detalhado da API de pedidos: regras de negocio, fluxos, rotas, persistencia, cache, mensageria, configuracao e testes.

## Visao Geral

O projeto e uma API backend de gerenciamento de pedidos de e-commerce. A aplicacao permite:

- Criar um pedido.
- Consultar um pedido por ID.
- Atualizar o status de um pedido.
- Consultar a saude das dependencias.
- Acessar documentacao interativa via Swagger.

O servico persiste pedidos no MongoDB, utiliza Redis como cache para consultas e publica eventos de alteracao de status no RabbitMQ.

## Fluxo de Inicializacao

`cmd/api/main.go` chama `app.New()`, registra `defer a.Close()` e executa `a.Run()`.

`internal/app/app.go` e o ponto de composicao da aplicacao:

1. Cria o logger Zap.
2. Carrega configuracoes com `config.LoadConfig`.
3. Conecta no MongoDB.
4. Cria o repositorio Mongo de pedidos.
5. Cria o client Redis e o adapter de cache.
6. Envolve o repositorio Mongo com o decorator de cache.
7. Conecta no RabbitMQ e declara filas default.
8. Instancia casos de uso.
9. Instancia handlers HTTP.
10. Registra healthcheck de MongoDB e RabbitMQ.
11. Cria o servidor Gin.

O fechamento de recursos acontece em `App.Close()`: MongoDB, Redis, RabbitMQ e logger.

## Dominio de Pedido

Entidade principal: `order.Order`.

Campos:

- `ID`: UUID string gerado na criacao.
- `CustomerID`: cliente dono do pedido.
- `Items`: lista de itens.
- `TotalAmount`: total em centavos.
- `Status`: estado atual do pedido.
- `CreatedAt`: data de criacao.
- `UpdatedAt`: data da ultima atualizacao.

Item:

- `ProductID`: identificador do produto.
- `Quantity`: quantidade.
- `Price`: preco unitario em centavos.

Status validos:

```text
created
processing
shipped
delivered
```

Transicoes validas:

```text
created -> processing -> shipped -> delivered
```

Transicoes fora dessa cadeia retornam `order.ErrInvalidStatusTransition`.

Erros de dominio:

- `ErrInvalidStatusTransition`
- `ErrOrderNotFound`
- `ErrEmptyOrder`
- `ErrInvalidItemPrice`

Regra financeira: dinheiro e sempre representado em centavos usando `int64`. Evite `float64`.

## Casos de Uso

### Criar Pedido

Arquivo: `internal/usecase/order/create_order.go`.

Responsabilidades:

- Receber `CreateOrderInput`.
- Converter itens de entrada para `order.OrderItem`.
- Criar entidade via `order.NewOrder`.
- Persistir usando `order.IOrderRepository`.

Validacoes:

- `customerId` obrigatorio no binding HTTP.
- `items` obrigatorio e com ao menos um item.
- `productId` obrigatorio.
- `quantity` maior que zero.
- `price` maior que zero no DTO HTTP/usecase.
- O dominio tambem rejeita lista vazia e preco negativo.

### Consultar Pedido

Arquivo: `internal/usecase/order/get_order.go`.

Responsabilidade:

- Buscar pedido por ID via `order.IOrderRepository`.

Na aplicacao real, a interface aponta para o repositorio decorado com cache.

### Atualizar Status

Arquivo: `internal/usecase/order/update_order_status.go`.

Responsabilidades:

- Buscar o pedido atual.
- Guardar status antigo.
- Validar transicao chamando `Order.UpdateStatus`.
- Persistir novo status no repositorio.
- Publicar evento `OrderStatusChangedEvent` na fila `order.status.changed`.

Observacao importante: a persistencia no MongoDB acontece antes da publicacao no RabbitMQ. Se o broker falhar apos o update, o banco fica atualizado e o caso de uso retorna erro. Nao ha Transactional Outbox.

## HTTP

Servidor: `internal/infra/http/server.go`.

Caracteristicas:

- Usa `gin.Default()`.
- CORS permissivo com `AllowOrigins: ["*"]`.
- Swagger em `/swagger/index.html`.
- `/swagger` redireciona para `/swagger/index.html`.
- Porta default: `3333`.

Rotas:

```text
POST  /orders
GET   /orders/:id
PATCH /orders/:id/status
GET   /health
GET   /swagger/index.html
```

### POST /orders

Cria um pedido.

Payload:

```json
{
  "customerId": "user-123",
  "items": [
    {
      "productId": "prod-abc",
      "quantity": 2,
      "price": 4990
    }
  ]
}
```

Retornos:

- `201`: pedido criado.
- `400`: JSON invalido ou falha de binding.
- `500`: erro ao criar ou salvar pedido.

Resposta de sucesso: objeto `order.Order`.

### GET /orders/:id

Busca um pedido por ID.

Retornos:

- `200`: pedido encontrado.
- `404`: pedido nao encontrado.
- `500`: erro inesperado.

Resposta de sucesso: objeto `order.Order`.

### PATCH /orders/:id/status

Atualiza o status de um pedido.

Payload:

```json
{
  "status": "processing"
}
```

Retornos:

- `200`: status atualizado.
- `400`: JSON invalido ou campo ausente.
- `404`: pedido nao encontrado.
- `422`: transicao de status invalida.
- `500`: erro inesperado, incluindo erro de publicacao no RabbitMQ.

Resposta de sucesso:

```json
{
  "message": "status updated successfully"
}
```

### GET /health

Executa healthcheck de dependencias registradas.

Dependencias atuais:

- MongoDB.
- RabbitMQ.

Retornos:

- `200`: todas as dependencias saudaveis.
- `503`: ao menos uma dependencia indisponivel.

Formato:

```json
{
  "status": "healthy",
  "services": {
    "mongodb": "up",
    "rabbitmq": "up"
  }
}
```

## Persistencia

Adapter MongoDB: `internal/infra/persistence/mongodb/order_repository.go`.

Comportamento:

- Database configurado por `DB_NAME`.
- Collection: `orders`.
- O ID do pedido e salvo no campo BSON `_id`.
- `Save` usa `InsertOne`.
- `GetByID` usa `FindOne` por `_id`.
- `mongo.ErrNoDocuments` vira `order.ErrOrderNotFound`.
- `UpdateStatus` usa `$set` para `status` e `$currentDate` para `updatedAt`.

Nao ha migrations nem criacao explicita de indices no projeto.

## Cache

Porta: `internal/domain/cache/interface.go`.

Adapter Redis:

- `internal/infra/cache/redis/connection.go`
- `internal/infra/cache/redis/cache.go`

Decorator de repositorio:

- `internal/infra/persistence/cached/order_repository.go`

Chave:

```text
order:{id}
```

TTL:

- Configurado por `ORDER_CACHE_TTL`.
- Default: `5m`.

Fluxo de consulta:

1. Tenta buscar no Redis.
2. Se encontrar valor valido, retorna do cache.
3. Se cache estiver vazio ou falhar, busca no MongoDB.
4. Depois de buscar no MongoDB, tenta popular o Redis.

Fluxo de criacao:

1. Salva no MongoDB.
2. Tenta popular cache.

Fluxo de atualizacao:

1. Atualiza status no MongoDB.
2. Tenta invalidar a chave no Redis.

Falhas de Redis nao devem interromper a operacao principal.

## Mensageria

Porta: `internal/domain/broker/interface.go`.

Adapter RabbitMQ:

- `internal/infra/broker/rabbitmq/message_broker.go`
- `internal/infra/broker/rabbitmq/connection.go`
- `internal/infra/broker/rabbitmq/queues.go`

Fila default:

```text
order.status.changed
```

Evento:

```go
type OrderStatusChangedEvent struct {
    OrderID   string
    OldStatus OrderStatus
    NewStatus OrderStatus
    Timestamp time.Time
}
```

Comportamento:

- Publica JSON.
- Usa default exchange.
- Usa o nome da fila como routing key.
- Declara filas no startup.
- Tenta reconectar em background se a conexao cair.

Startup:

- RabbitMQ e dependencia obrigatoria.
- A conexao tenta ate 10 vezes.
- Intervalo entre tentativas: 3 segundos.

## Configuracao

Arquivo de exemplo: `.env.example`.

Variaveis:

```text
PORT=3333
MONGO_USER=admin
MONGO_PASS=admin123
MONGO_URI=mongodb://admin:admin123@mongodb:27017/
DB_NAME=order-api
RABBITMQ_USER=guest
RABBITMQ_PASS=guest
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
ORDER_CACHE_TTL=5m
```

Defaults no codigo:

- `PORT`: `3333`
- `MONGO_URI`: `mongodb://localhost:27017`
- `DB_NAME`: `order-api`
- `RABBITMQ_URL`: `amqp://guest:guest@localhost:5672/`
- `REDIS_ADDR`: `localhost:6379`
- `REDIS_PASSWORD`: vazio
- `REDIS_DB`: `0`
- `ORDER_CACHE_TTL`: `5m`

`config.LoadConfig` carrega `.env` se existir e usa variaveis do ambiente do processo como fonte.

Observacao: `Server.Start()` le `PORT` diretamente com `os.Getenv`, enquanto `config.LoadConfig` tambem popula `AppConfig.Port`.

## Docker

Subir ambiente completo:

```bash
docker compose build --no-cache
docker compose up -d
```

Servicos:

- `app`: API na porta `3333`.
- `mongodb`: MongoDB `8.0`, porta `27017`.
- `rabbitmq`: RabbitMQ `4.0-management`, portas `5672` e `15672`.
- `redis`: Redis `8.2-alpine`, porta `6379`.

RabbitMQ Management:

```text
http://localhost:15672
```

Dockerfile:

- Build multi-stage com `golang:1.26.1-alpine`.
- Instala `swag`.
- Executa `swag init -g cmd/api/main.go --parseDependency --parseInternal`.
- Compila `./cmd/api/main.go`.
- Imagem final baseada em Alpine.

## Swagger

Arquivos gerados:

- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

Regenerar:

```bash
swag init -g cmd/api/main.go --parseDependency --parseInternal
```

Ao alterar handlers, payloads, status HTTP ou comentarios Swagger, regenere os arquivos.

## Testes

Cobertura atual relevante:

- Dominio: `internal/domain/order`.
- Use cases: `internal/usecase/order`.
- Handlers HTTP: `internal/infra/http/handler`.
- Decorator de cache: `internal/infra/persistence/cached`.

Comando geral:

```bash
go test ./...
```

Comando com cobertura para areas principais:

```bash
go test -v ./internal/domain/order/... ./internal/usecase/order/... ./internal/infra/http/handler/... ./internal/infra/persistence/cached/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Ao alterar regras de status:

- Atualize `internal/domain/order/entity_test.go`.
- Atualize `internal/usecase/order/update_order_status_test.go`.
- Revise testes de handler se o mapeamento HTTP mudar.

Ao alterar cache:

- Atualize `internal/infra/persistence/cached/order_repository_test.go`.

Ao alterar contratos HTTP:

- Atualize `internal/infra/http/handler/order_test.go`.
- Atualize anotacoes Swagger.
- Regenere `docs`.

## Exemplos de Uso

Criar pedido:

```bash
curl -X POST http://localhost:3333/orders \
  -H 'Content-Type: application/json' \
  -d '{
    "customerId": "user-123",
    "items": [
      {
        "productId": "prod-abc",
        "quantity": 2,
        "price": 4990
      }
    ]
  }'
```

Atualizar status:

```bash
curl -X PATCH http://localhost:3333/orders/<order-id>/status \
  -H 'Content-Type: application/json' \
  -d '{"status":"processing"}'
```

Consultar pedido:

```bash
curl http://localhost:3333/orders/<order-id>
```

Healthcheck:

```bash
curl http://localhost:3333/health
```

## Melhorias Futuras Ja Identificadas

- Transactional Outbox para consistencia entre MongoDB e RabbitMQ.
- Autenticacao e autorizacao.
- CORS restrito por ambiente.
- Metricas com Prometheus/Grafana.
- Tracing distribuido com OpenTelemetry.
- Circuit breaker para dependencias externas.
- Idempotencia em criacao e atualizacao.
- Migrations/indices para MongoDB.
- Consumers para eventos publicados.

