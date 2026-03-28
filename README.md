# Order Management Service

Este é um serviço backend de gerenciamento de pedidos de e-commerce, desenvolvido em Go utilizando a arquitetura Hexagonal (Ports and Adapters).

## Stack Utilizada

- **Linguagem:** Go 1.26.1
- **Framework Web:** [Gin](https://gin-gonic.com/)
- **Banco de Dados:** [MongoDB](https://www.mongodb.com/)
- **Mensageria:** [RabbitMQ](https://www.rabbitmq.com/)
- **Logger:** [Zap (Uber)](https://github.com/uber-go/zap)
- **Containerização:** Docker & Docker Compose
- **Testes:** [Testify](https://github.com/stretchr/testify)

## Arquitetura

O projeto segue os princípios da **Arquitetura Hexagonal**:
- **Domain:** Definições core de negócio (`Order`), eventos e interfaces agnósticas (`IMessageBroker`, `Logger`, `HealthChecker`).
- **Application (Use Cases):** Orquestra a lógica de negócio de forma isolada da infraestrutura.
- **Infrastructure (Adapters):** Implementações concretas de persistência (MongoDB), mensageria (RabbitMQ), monitoramento de saúde e servidor HTTP (Gin).

### Destaques da Implementação
- **Broker Genérico:** O sistema de mensageria é totalmente desacoplado, permitindo publicar qualquer payload em qualquer fila/tópico através de uma interface genérica.
- **Healthcheck Dinâmico:** Um sistema de monitoramento extensível que valida a saúde de múltiplos serviços (MongoDB, RabbitMQ) de forma independente.
- **Logging Estruturado:** Injeção de dependência do Logger em todas as camadas, permitindo fácil substituição e rastreabilidade.
- **Precisão Financeira:** Utilização de `int64` para valores monetários (centavos) para evitar problemas de precisão com pontos flutuantes.

## Como Executar

Para subir toda a infraestrutura e a aplicação:

```bash
docker compose build --no-cache && docker compose up -d
```

A aplicação estará disponível em `http://localhost:3333`.

## Endpoints

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/orders` | Criar novo pedido |
| GET | `/orders/:id` | Consultar pedido por ID |
| PATCH | `/orders/:id/status` | Atualizar status do pedido |
| GET | `/health` | Healthcheck detalhado (Mongo + Rabbit) |
| GET | `/swagger/index.html` | Documentação interativa (Swagger UI) |

A documentação interativa pode ser acessada em `http://localhost:3333/swagger/index.html`.

### Exemplo de criação de pedido (`POST /orders`):

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
*Nota: O `price` deve ser enviado em centavos (Ex: 49.90 -> 4990).*

## Testes

Os testes unitários cobrem os casos de uso principais e as regras de transição de status utilizando mocks para as interfaces de infraestrutura.

Para executar os testes focados na lógica de negócio (Entidades, Casos de Uso e Handlers) e ver a cobertura:

```bash
go test -v ./internal/domain/order/... ./internal/usecase/order/... ./internal/infra/http/handler/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Variáveis de Ambiente

O arquivo `.env` (baseado no `.env.example`) controla as configurações:
- `PORT`: Porta do servidor HTTP (default: 3333).
- `MONGO_URI`: URI de conexão com o MongoDB.
- `MONGO_USER`: Usuário do MongoDB.
- `MONGO_PASS`: Senha do MongoDB.
- `DB_NAME`: Nome do banco de dados.
- `RABBITMQ_URL`: URL de conexão com o RabbitMQ.
- `RABBITMQ_USER`: Usuário do RabbitMQ.
- `RABBITMQ_PASS`: Senha do RabbitMQ.

## 🚀 Melhorias Futuras (Roadmap)

Embora o projeto já siga boas práticas de arquitetura, as seguintes evoluções são recomendadas para um ambiente de produção real:

1.  **Validações de Domínio Estendidas**:
    *   **Estoque**: Integração com um serviço de catálogo/estoque para validar a disponibilidade antes de confirmar o pedido.
    *   **Clientes**: Validação de existência e status do cliente em um serviço de Identity/CRM.

2.  **Resiliência e Consistência (Transactional Outbox)**:
    *   Implementar o **Outbox Pattern** para garantir a consistência eventual. Em vez de publicar diretamente no RabbitMQ, os eventos seriam salvos no MongoDB na mesma transação do pedido e um worker separado processaria o envio com **Retry e Exponential Backoff**.

3.  **Seguraça**:
    *   Implementação de autenticação e autorização (ex: JWT/OAuth2) para garantir que apenas clientes e sistemas autorizados acessem os endpoints.

4.  **Observabilidade**:
    *   **Métricas**: Integração com Prometheus/Grafana para monitorar latência, taxas de erro e vazão.
    *   **Tracing**: Implementação de OpenTelemetry para rastreio distribuído de requisições entre camadas e serviços.

5.  **Estabilidade**:
    *   **Circuit Breaker**: Implementação de disjuntores (ex: `gobreaker`) nas chamadas para RabbitMQ e MongoDB para evitar falhas em cascata durante instabilidades na infraestrutura.

6.  **Idempotência**:
    *   Implementação de suporte a chaves de idempotência (ex: header `X-Idempotency-Key`) nos endpoints de criação e atualização para evitar o processamento duplicado de requisições em caso de falhas de rede ou retentativas automáticas do cliente.
