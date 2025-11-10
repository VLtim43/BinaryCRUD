# Relatório Técnico - TP III

## Formulário Técnico

### 1. Qual foi o relacionamento N:N escolhido e quais tabelas ele conecta?

**Relacionamento:** Orders ↔ Promotions

**Justificativa:**

- Um pedido (Order) pode ter múltiplas promoções aplicadas
- Uma promoção (Promotion) pode ser aplicada a múltiplos pedidos

**Tabela Intermediária:** `order_promotions.bin`

**Entidades conectadas:**

- `orders.bin` - Pedidos do restaurante (ID, customer name, total price, item IDs)
- `promotions.bin` - Promoções disponíveis (ID, promotion name, total price, item IDs)

**Localização:** `backend/dao/order_promotion_dao.go`

---

### 2. Qual estrutura de índice foi utilizada (B+ ou Hash Extensível)? Justifique a escolha.

**Estrutura:** **Nenhum índice** (scan sequencial)

**Justificativa:**

A tabela intermediária `order_promotions.bin` **não utiliza índice** pelas seguintes razões:

1. **Volume de dados reduzido:** Relacionamentos N:N tendem a ter menos registros que as tabelas principais

   - Exemplo: 100 orders × 5 promotions = ~500 registros máximo

2. **Padrão de acesso:** Buscas são sempre por chave completa ou parcial

   - `GetByOrderID(orderID)` - busca todos os relacionamentos de um pedido
   - `GetByPromotionID(promotionID)` - busca todos os relacionamentos de uma promoção

3. **Simplicidade e consistência:**

   - Orders e Promotions já usam scan sequencial
   - Manter o mesmo padrão reduz complexidade
   - Menor overhead de manutenção (sem persistência de índice)

---

### 3. Como foi implementada a chave composta da tabela intermediária?

**Formato binário da chave composta:**

```
[orderID(2)][0x1F][promotionID(2)][0x1F][tombstone(1)][0x1E]
```

**Características:**

1. **Sem ID auto-incremental:** A chave primária **É** a combinação (orderID, promotionID)
2. **Ordem determinística:** orderID sempre vem primeiro, seguido de promotionID
3. **Tamanho fixo:** 2 bytes para cada ID da composição

**Implementação:**

```go
type OrderPromotion struct {
    OrderID     uint64
    PromotionID uint64
}
```

**Unicidade garantida:**

A função `existsUnlocked(orderID, promotionID)` verifica duplicatas antes de inserir:

```go
func (dao *OrderPromotionDAO) Write(orderID, promotionID uint64) error {
    // Check for duplicate composite key
    exists, err := dao.existsUnlocked(orderID, promotionID)
    if exists {
        return fmt.Errorf("relationship already exists")
    }
    // ... write entry
}
```

**Processo de verificação:**

1. Scan sequencial de todos os registros
2. Parse de cada entry para extrair (orderID, promotionID)
3. Verifica se tombstone == 0x00 (ativo)
4. Compara ambos os campos da chave composta
5. Retorna erro se duplicata encontrada

**Localização:** `backend/dao/order_promotion_dao.go:154-218` (existsUnlocked)

---

### 4. Como é feita a busca eficiente de registros por meio do índice?

**Método:** **Scan sequencial com filtro** (sem índice)

Duas operações de busca são fornecidas:

#### a) GetByOrderID - Buscar promoções de um pedido

```go
func (dao *OrderPromotionDAO) GetByOrderID(orderID uint64) ([]*OrderPromotion, error)
```

**Algoritmo:**

1. Lê todo o arquivo em memória
2. Divide por record separator (`0x1E`)
3. Para cada entrada:
   - Parse orderID, promotionID, tombstone
   - **Filtra:** Se `tombstone == 0x00 && entryOrderID == orderID`
   - Adiciona ao resultado
4. Retorna array de OrderPromotion

**Complexidade:** O(n) onde n = número total de relacionamentos

#### b) GetByPromotionID - Buscar pedidos com uma promoção

```go
func (dao *OrderPromotionDAO) GetByPromotionID(promotionID uint64) ([]*OrderPromotion, error)
```

**Algoritmo:** Idêntico ao anterior, mas filtra por `entryPromotionID == promotionID`

**Otimizações aplicadas:**

- **Early skip:** Registros deletados (tombstone != 0x00) são ignorados imediatamente
- **Single I/O:** Um único `os.ReadFile()` carrega todos os dados
- **Parse eficiente:** Offset tracking evita cópias de memória desnecessárias

**Localização:**

- `backend/dao/order_promotion_dao.go:220-304` (GetByOrderID)
- `backend/dao/order_promotion_dao.go:306-390` (GetByPromotionID)

**Trade-off:** O(n) é aceitável para volumes pequenos. Para grandes volumes, considera-se adicionar índice Hash ou B+ Tree.

---

### 5. Como o sistema trata a integridade referencial (remoção/atualização) entre as tabelas?

**Estratégia:** **Sem cascading operations**

#### Cenários de Integridade:

**a) Criação de relacionamento:**

✅ **Sem validação de existência:**

- Sistema **não valida** se orderID ou promotionID existem ao criar relacionamento
- Permite relacionamentos "orfãos" temporários
- Responsabilidade da camada de aplicação validar antes de chamar Write()

```go
// Aplicação deve validar:
order, err := orderDAO.Read(orderID)
promo, err := promotionDAO.Read(promotionID)
// Só então criar relacionamento
orderPromoDAO.Write(orderID, promotionID)
```

**b) Deleção de Order:**

❌ **Sem cascading delete:**

- Deletar order com `orderDAO.Delete(orderID)` **não remove** relacionamentos
- Relacionamentos órfãos permanecem em `order_promotions.bin`
- Leitura via `GetByOrderID(orderID)` retorna relacionamentos, mas order não existe mais

**c) Deleção de Promotion:**

❌ **Sem cascading delete:**

- Idêntico ao cenário anterior
- `GetByPromotionID(promotionID)` retorna relacionamentos órfãos

**d) Deleção de relacionamento:**

✅ **Suportado:**

```go
func (dao *OrderPromotionDAO) Delete(orderID, promotionID uint64) error
```

- Marca tombstone = 0x01 no relacionamento específico
- Não afeta orders ou promotions

#### Justificativa da Estratégia:

**Vantagens:**

- **Performance:** Evita lookups custosos durante deletes
- **Simplicidade:** Sem lógica de cascading complexa
- **Flexibilidade:** Aplicação decide política de integridade

**Desvantagens:**

- Dados órfãos possíveis
- Requer limpeza periódica (garbage collection)
- Aplicação deve validar integridade

---

### 6. Como foi organizada a persistência dos dados dessa nova tabela (mesmo padrão de cabeçalho e lápide)?

**Formato:** Mesmo padrão das tabelas principais

#### Header (15 bytes):

```
[entitiesCount(4)][0x1F][tombstoneCount(4)][0x1F][nextId(4)][0x1E]
```

- `entitiesCount`: Número de relacionamentos ativos
- `tombstoneCount`: Número de relacionamentos deletados
- `nextId`: **Não utilizado** (não há auto-increment, chave é composta)

#### Registro de OrderPromotion:

```
[orderID(2)][0x1F][promotionID(2)][0x1F][tombstone(1)][0x1E]
```

**Campos:**

- `orderID` (2 bytes): ID do pedido
- `0x1F`: Unit separator
- `promotionID` (2 bytes): ID da promoção
- `0x1F`: Unit separator
- `tombstone` (1 byte): 0x00 = ativo, 0x01 = deletado
- `0x1E`: Record separator (fim do registro)

#### Exclusão Lógica:

**Processo de Delete:**

1. Scan sequencial até encontrar (orderID, promotionID)
2. Verifica se tombstone já é 0x01 (já deletado)
3. Calcula posição do tombstone no arquivo:
   ```
   position = entryStart + 2 (orderID) + 1 (sep) + 2 (promotionID) + 1 (sep)
   ```
4. `file.Seek(position)` para posição exata
5. `file.Write([]byte{0x01})` marca como deletado
6. `file.Sync()` força escrita em disco
7. Atualiza header: `entitiesCount--`, `tombstoneCount++`

**Consistência:**

- Sync duplo: primeiro dados, depois header
- Se falha entre syncs, dado é deletado mas contador inconsistente (recuperável)

**Localização:**

- `backend/dao/order_promotion_dao.go:26-52` (ensureFileExists)
- `backend/dao/order_promotion_dao.go:54-151` (Write)
- `backend/dao/order_promotion_dao.go:392-516` (Delete)

---

### 7. Descreva como o código da tabela intermediária se integra com o CRUD das tabelas principais.

**Arquitetura:** Camada de abstração independente

#### Estrutura de Integração:

```
app.go
   ↓
OrderPromotionDAO ← independente → OrderDAO / PromotionDAO
   ↓
order_promotions.bin
```

**Características:**

1. **Independência total:** OrderPromotionDAO não depende de OrderDAO ou PromotionDAO
2. **Sem validação automática:** Aplicação valida existência manualmente
3. **API pública:** Métodos expostos via Wails bindings

#### Métodos Expostos em app.go:

```go
// Create relationship
func (a *App) CreateOrderPromotion(orderID, promotionID int) error {
    return a.orderPromoDAO.Write(uint64(orderID), uint64(promotionID))
}

// Get promotions for an order
func (a *App) GetPromotionsByOrderID(orderID int) ([]*dao.OrderPromotion, error) {
    return a.orderPromoDAO.GetByOrderID(uint64(orderID))
}

// Get orders with a promotion
func (a *App) GetOrdersByPromotionID(promotionID int) ([]*dao.OrderPromotion, error) {
    return a.orderPromoDAO.GetByPromotionID(uint64(promotionID))
}

// Delete relationship
func (a *App) DeleteOrderPromotion(orderID, promotionID int) error {
    return a.orderPromoDAO.Delete(uint64(orderID), uint64(promotionID))
}
```

#### Fluxo de Uso Típico (Frontend):

```typescript
// 1. Validar existência (aplicação deve fazer)
const order = await GetOrderByID(orderId);
const promo = await GetPromotionByID(promoId);

// 2. Criar relacionamento
if (order && promo) {
  await CreateOrderPromotion(orderId, promoId);
}

// 3. Listar promoções de um pedido
const promos = await GetPromotionsByOrderID(orderId);
for (const rel of promos) {
  const promotion = await GetPromotionByID(rel.PromotionID);
  console.log(promotion.OwnerOrName);
}
```

#### Padrão DAO:

- **Isolamento:** Cada DAO gerencia seu próprio arquivo
- **Mutex:** Thread-safety dentro de cada DAO
- **Sem transações:** Operações não são atômicas entre tabelas
- **Fail-fast:** Erros propagados até app.go

**Localização:**

- `app.go:60-61` (Inicialização do DAO)
- `app.go` (Métodos API não mostrados no código fornecido, mas estrutura presumida)

---

### 8. Descreva como está organizada a estrutura de diretórios e módulos no repositório após esta fase.

**Estrutura atualizada:**

```
BinaryCRUD/
├── backend/
│   ├── dao/                      # Data Access Objects
│   │   ├── item_dao.go           # CRUD de items com B+ Tree index
│   │   ├── collection_dao.go     # Lógica compartilhada orders/promotions
│   │   ├── order_dao.go          # Wrapper para orders
│   │   ├── promotion_dao.go      # Wrapper para promotions
│   │   └── order_promotion_dao.go   # 🆕 Tabela intermediária N:N
│   ├── index/                    # Estrutura de indexação
│   │   ├── btree.go              # B+ Tree implementation
│   │   └── persistence.go        # Serialização de índices
│   ├── utils/                    # Utilitários binários
│   │   ├── write.go              # Escrita binária
│   │   ├── read.go               # Leitura binária
│   │   ├── header.go             # Gerenciamento de headers
│   │   ├── finder.go             # Busca sequencial por ID
│   │   ├── constants.go          # Constantes (separadores, tamanhos)
│   │   └── file.go               # Operações de arquivo
│   └── test/                     # Testes unitários
│       ├── item_dao_test.go      # Testes de items
│       ├── collection_dao_test.go # Testes de collections
│       ├── order_dao_test.go     # 🆕 Testes de orders
│       ├── promotion_dao_test.go # 🆕 Testes de promotions
│       ├── btree_test.go         # Testes de B+ Tree
│       ├── read_test.go          # Testes de leitura binária
│       ├── write_test.go         # Testes de escrita binária
│       └── file_test.go          # Testes de operações de arquivo
├── frontend/
│   ├── src/
│   │   ├── app.tsx               # Aplicação Preact
│   │   └── App.scss              # Estilos
│   └── wailsjs/                  # Bindings Wails auto-gerados
├── data/                         # Persistência binária
│   ├── items.bin                 # Registros de items
│   ├── items.idx                 # Índice B+ Tree de items
│   ├── orders.bin                # Registros de orders
│   ├── promotions.bin            # Registros de promotions
│   └── order_promotions.bin      # 🆕 Tabela intermediária N:N
├── logs/
│   └── app.log                   # Logs da aplicação
├── app.go                        # API backend (Wails bindings)
├── main.go                       # Entry point
├── logger.go                     # Sistema de logging
├── relatorioTPII.md              # Relatório fase anterior
├── relatorioTPIII.md             # 🆕 Este relatório
└── README.md                     # Documentação

```

---

## Observações Técnicas

### Limitações Identificadas:

1. **🚨 Bug crítico:** Valores numéricos contendo bytes `0x1E` ou `0x1F` causam corrupção de dados

   - Exemplo: preço 798 (0x031E) quebra parsing
   - **Solução temporária:** Evitar valores com esses bytes
   - **Solução definitiva:** Implementar escaping ou usar length-prefixed encoding

2. **Sem índices em Orders/Promotions:** Lookups são O(n)

   - Aceitável para volumes pequenos
   - Requer otimização se escalar

### Melhorias Futuras:

- Adicionar Hash index em order_promotions para O(1) lookups
- Implementar garbage collection de registros tombstoned
- Adicionar validação de foreign keys opcional
- Migrar para encoding length-prefixed (evitar bug de separadores)
