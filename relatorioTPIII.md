# Relatório Técnico - TP III

## Formulário Técnico

### 1. Qual foi o relacionamento N:N escolhido e quais tabelas ele conecta?

**Relacionamento:** Orders ↔ Promotions

- Um pedido (Order) pode ter múltiplas promoções aplicadas
- Uma promoção (Promotion) pode ser aplicada a múltiplos pedidos

**Tabela Intermediária:** `order_promotions.bin`

**Entidades conectadas:**

- `orders.bin` - Pedidos do restaurante (ID, customer name, total price, item IDs)
- `promotions.bin` - Promoções disponíveis (ID, promotion name, total price, item IDs)


---

### 2. Qual estrutura de índice foi utilizada (B+ ou Hash Extensível)? Justifique a escolha.

**Estrutura:** **B+ Tree** (ordem 4)

**Aplicado em:**
- `items.bin` → `items.idx`
- `orders.bin` → `orders.idx`
- `promotions.bin` → `promotions.idx`

**Não indexado:**
- `order_promotions.bin` (tabela intermediária N:N)

**Justificativa para B+ Tree:**

1. **Range queries eficientes:** B+ Tree mantém dados ordenados nos nós folha
2. **Operações balanceadas:** Insert/Search/Delete em O(log n)
3. **Cache-friendly:** Nós folha ligados em lista para scan sequencial
4. **Persistência simples:** Estrutura serializa facilmente em formato binário
5. **Baixo overhead:** Ordem 4 mantém árvore balanceada sem muitos nós

**Justificativa para não indexar order_promotions:**

1. **Volume reduzido:** Relacionamentos N:N têm menos registros que entidades principais
2. **Buscas sempre filtradas:** Queries são sempre por orderID ou promotionID completo
3. **Simplicidade:** Scan sequencial suficiente para datasets pequenos

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

---

### 4. Como é feita a busca eficiente de registros por meio do índice?

**Método:** **Scan sequencial com filtro** (sem índice)

Duas operações de busca são fornecidas:

#### a) GetByOrderID - Buscar promoções de um pedido
#### b) GetByPromotionID - Buscar pedidos com uma promoção

---

### 5. Como o sistema trata a integridade referencial (remoção/atualização) entre as tabelas?

**Estratégia:** **Sem cascading operations**

#### Cenários de Integridade:

**a) Criação de relacionamento:**

✅ **Sem validação de existência:**

- Sistema **não valida** se orderID ou promotionID existem ao criar relacionamento
- Permite relacionamentos "orfãos" temporários
- Responsabilidade da camada de aplicação validar antes de chamar Write()

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

- Marca tombstone = 0x01 no relacionamento específico
- Não afeta orders ou promotions

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
│   ├── orders.idx                # 🆕 Índice B+ Tree de orders
│   ├── promotions.bin            # Registros de promotions
│   ├── promotions.idx            # 🆕 Índice B+ Tree de promotions
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

2. **✅ RESOLVIDO:** Orders e Promotions agora usam B+ Tree indexing (O(log n) lookups)

### Melhorias Futuras:

- Adicionar Hash index em order_promotions para O(1) lookups
- Implementar garbage collection de registros tombstoned
- Adicionar validação de foreign keys opcional
- Migrar para encoding length-prefixed (evitar bug de separadores)
