# AI Development Guidelines

Este projeto segue Arquitetura Hexagonal (Ports and Adapters).

A IA deve seguir estritamente estes padrões ao criar ou modificar código.

---

## 🧱 Arquitetura

- Padrão: Hexagonal (Ports and Adapters)
- Separação de camadas obrigatória:

  - `internal/domain` → regras de negócio puras (NÃO depende de infra)
  - `internal/usecase` → orquestra regras de negócio
  - `internal/infra` → implementações externas (DB, HTTP, cache, broker, etc)
  - `internal/app` → composição das dependências
  - `cmd/*` → entrypoints da aplicação

### Regras fundamentais

- ❌ Domain NÃO pode importar bibliotecas externas (frameworks, DBs, etc)
- ❌ Domain NÃO pode depender de infra

- ✅ Domain define interfaces (ports)
- ✅ Infra implementa interfaces (adapters)
- ✅ Usecases dependem apenas de interfaces do domínio

---

## 🧠 Regras de Domínio

- O domínio deve ser:
  - puro
  - independente
  - testável sem dependências externas

### Boas práticas

- Entidades devem encapsular regras de negócio
- Evitar structs anêmicas (sem comportamento)
- Validações críticas devem estar no domínio

### Dinheiro

- Sempre usar `int64` (centavos ou menor unidade)
- ❌ Nunca usar `float64`

---

## ⚙️ Stack e Tecnologias

- Linguagem: Go
- HTTP: Gin (ou equivalente)
- Banco: qualquer persistência (ex: MongoDB, Postgres)
- Cache: Redis (opcional)
- Mensageria: RabbitMQ/Kafka (ou equivalente)
- Logger: estruturado (ex: Zap) via interface
- Config: variáveis de ambiente
- Docs: Swagger/OpenAPI

---

## 🧩 Padrões de Implementação

### Repositórios

- Definidos no domínio como interfaces
- Implementados na camada de infraestrutura

Exemplo:

```go
type Repository interface {
    Save(ctx context.Context, entity *Entity) error
    GetByID(ctx context.Context, id string) (*Entity, error)
}
````

---

### Cache

* Sempre tratar como **best effort**
* ❌ Nunca quebrar fluxo principal se falhar
* ❌ Nunca ser fonte de verdade

Padrões recomendados:

* read-through
* write + invalidate

---

### Mensageria

* Publicação de eventos deve ocorrer após persistência

⚠️ Importante:

* Sem transactional outbox por padrão
* Pode haver inconsistência entre banco e eventos
* Falha na publicação deve retornar erro

---

### Logging

* Usar interface (ex: `logger.Logger`)
* ❌ Não acoplar diretamente a biblioteca (ex: Zap)

---

### Context

* Sempre propagar `context.Context` entre camadas

---

## 🌐 HTTP Layer

### Responsabilidades

* Converter request → DTO
* Chamar usecase
* Mapear erros → status HTTP

### Boas práticas

* Não conter regra de negócio
* Não acessar diretamente infra

---

## 🧪 Testes

Ao alterar código:

* Atualizar testes de domínio
* Atualizar testes de usecase
* Atualizar testes de adapters/handlers

Ferramentas comuns:

* `testing`
* `testify`

---

## ⚠️ Cuidados Importantes

* Banco de dados é fonte de verdade
* Cache deve ser opcional
* Mensageria deve ser tratada como dependência externa (falhas possíveis)

---

## 🚀 Ao Criar Novas Features

A IA deve seguir:

1. Definir ou ajustar regra no `domain`
2. Criar/ajustar usecase
3. Usar interfaces do domínio
4. Implementar adapters em `infra` (se necessário)
5. Injetar dependências no composition root (`app`)
6. Criar/ajustar handlers/interfaces de entrada
7. Adicionar testes
8. Atualizar documentação se necessário

---

## ❌ Anti-patterns (NÃO FAZER)

* Colocar lógica de negócio fora do domínio
* Acessar banco diretamente no usecase
* Usar cache como fonte de verdade
* Ignorar `context.Context`
* Usar `float` para valores monetários
* Acoplar domínio a frameworks ou libs externas
* Misturar responsabilidades entre camadas

---

## 📌 Princípios

* Clareza > abstração desnecessária
* Simplicidade > overengineering
* Consistência > inovação isolada
* Baixo acoplamento entre camadas
* Alta coesão dentro das camadas
