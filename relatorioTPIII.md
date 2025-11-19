# Relatório Técnico - TP III

## Formulário Técnico

### 1. Qual foi o relacionamento N:N escolhido e quais tabelas ele conecta?

**Relacionamento:** Orders ↔ Promotions

- Um pedido pode ter múltiplas promoções aplicadas
- Uma promoção pode ser aplicada a múltiplos pedidos

**Tabela Intermediária:** `order_promotions.bin`

**Entidades conectadas:**

- `orders.bin` - Pedidos (ID, cliente, preço total, IDs dos items)
- `promotions.bin` - Promoções (ID, nome, preço total, IDs dos items)

---

### 2. Qual estrutura de índice foi utilizada (B+ ou Hash Extensível)? Justifique a escolha.

**Estrutura:** B+ Tree (ordem 4)

**Aplicado em:**

- `items.bin` → `items.idx`
- `orders.bin` → `orders.idx`
- `promotions.bin` → `promotions.idx`

**Não indexado:**

- `order_promotions.bin` (tabela intermediária N:N)

**Justificativa:**

A B+ Tree foi escolhida por manter dados ordenados nos nós folha, permitindo range queries eficientes. Operações de Insert/Search/Delete são O(log n) e a estrutura serializa facilmente em formato binário. A tabela intermediária não foi indexada pois tem volume reduzido e buscas sempre filtradas por orderID ou promotionID completo.

---

### 3. Como foi implementada a chave composta da tabela intermediária?

**Formato binário:**

```text
[recordLength(2)][orderID(2)][promotionID(2)][tombstone(1)]
```

A chave primária é a combinação (orderID, promotionID). Não há ID auto-incremental - a própria combinação identifica unicamente o relacionamento. OrderID sempre vem primeiro, seguido de promotionID, cada um com 2 bytes. O prefixo de tamanho (2 bytes) indica o comprimento do registro.

---

### 4. Como é feita a busca eficiente de registros por meio do índice?

**Para entidades principais (items, orders, promotions):**

A B+ Tree mapeia ID → offset no arquivo binário. A busca percorre a árvore em O(log n) até encontrar o nó folha com o ID desejado, então faz seek direto no arquivo.

**Para tabela intermediária:**

Scan sequencial com filtro por orderID ou promotionID. Como o volume é reduzido, essa abordagem é suficiente.

---

### 5. Como o sistema trata a integridade referencial (remoção/atualização) entre as tabelas?

**Estratégia:** Sem cascading operations

**Criação de relacionamento:**

- Valida se order e promotion existem antes de criar o relacionamento
- Retorna erro se algum não existir

**Deleção de Order ou Promotion:**

- Não remove automaticamente os relacionamentos
- Relacionamentos órfãos permanecem em `order_promotions.bin`

**Deleção de relacionamento:**

- Marca tombstone = 0x01 no registro específico
- Não afeta orders ou promotions

---

### 6. Como foi organizada a persistência dos dados dessa nova tabela (mesmo padrão de cabeçalho e lápide)?

**Header (12 bytes):**

```text
[entitiesCount(4)][tombstoneCount(4)][nextId(4)]
```

- `entitiesCount`: Número de relacionamentos ativos
- `tombstoneCount`: Número de relacionamentos deletados
- `nextId`: Não utilizado (chave é composta)

**Registro:**

```text
[recordLength(2)][orderID(2)][promotionID(2)][tombstone(1)]
```

Mesmo padrão das tabelas principais com prefixo de tamanho (length-prefixed encoding).

---

### 7. Descreva como o código da tabela intermediária se integra com o CRUD das tabelas principais

**Arquitetura:** Camada independente

O `OrderPromotionDAO` opera de forma independente dos outros DAOs. A integração acontece na camada de aplicação (`app.go`), que:

1. Valida existência de order e promotion antes de criar relacionamento
2. Busca relacionamentos e enriquece com dados das entidades
3. Calcula preço combinado (items + promotions) ao retornar pedido

O frontend chama `ApplyPromotionToOrder` para criar relacionamentos e `GetOrderWithPromotions` para obter pedido com suas promoções.

---

### 8. Descreva como está organizada a estrutura de diretórios e módulos no repositório após esta fase

```text
BinaryCRUD/
├── backend/
│   ├── dao/
│   │   ├── item_dao.go
│   │   ├── collection_dao.go
│   │   ├── order_dao.go
│   │   ├── promotion_dao.go
│   │   └── order_promotion_dao.go
│   ├── index/
│   │   ├── btree.go
│   │   └── persistence.go
│   ├── utils/
│   │   ├── constants.go
│   │   ├── dao_helper.go
│   │   ├── deleter.go
│   │   ├── entry.go
│   │   ├── file.go
│   │   ├── finder.go
│   │   ├── header.go
│   │   ├── parser.go
│   │   ├── read.go
│   │   └── write.go
│   └── test/
│       ├── app_test.go
│       ├── btree_test.go
│       ├── collection_dao_test.go
│       ├── dao_helper_test.go
│       ├── deleter_test.go
│       ├── file_test.go
│       ├── item_dao_test.go
│       ├── order_dao_test.go
│       ├── order_promotion_dao_test.go
│       ├── parser_test.go
│       ├── promotion_dao_test.go
│       ├── read_test.go
│       └── write_test.go
├── frontend/
│   └── src/
│       ├── app.tsx
│       ├── App.scss
│       ├── components/
│       │   ├── Button.tsx
│       │   ├── DataTable.tsx
│       │   ├── Input.tsx
│       │   ├── ItemList.tsx
│       │   ├── LogsPanel.tsx
│       │   ├── Modal.tsx
│       │   ├── Select.tsx
│       │   └── tabs/
│       │       ├── ItemTab.tsx
│       │       ├── OrderTab.tsx
│       │       ├── PromotionTab.tsx
│       │       └── DebugTab.tsx
│       ├── services/
│       │   ├── itemService.ts
│       │   ├── orderService.ts
│       │   ├── promotionService.ts
│       │   ├── orderPromotionService.ts
│       │   ├── systemService.ts
│       │   └── logService.ts
│       ├── types/
│       │   └── cart.ts
│       └── utils/
│           └── formatters.ts
├── data/
│   ├── items.bin
│   ├── items.idx
│   ├── orders.bin
│   ├── orders.idx
│   ├── promotions.bin
│   ├── promotions.idx
│   └── order_promotions.bin
├── app.go
├── main.go
└── logger.go
```

---

## Observações Técnicas

### Limitações

1. Sem garbage collection de registros deletados (tombstones acumulam)
2. Sem cascading delete nos relacionamentos

### Melhorias Futuras

- Adicionar garbage collection periódica
- Implementar índice hash em order_promotions para lookups O(1)
