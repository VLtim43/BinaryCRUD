# Relatório Técnico - TP II

## a) Estrutura dos Registros

**Formato binário com campos fixos e variáveis:**

**Header (15 bytes):**

```
[entitiesCount(4)][0x1F][tombstoneCount(4)][0x1F][nextId(4)][0x1E]
```

**Registro de Item:**

```
[ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][price(4)][0x1E]
```

**Registro de Order/Promotion:**

```
[ID(2)][tombstone(1)][0x1F][nameSize(2)][name][0x1F][totalPrice(4)][0x1F][itemCount(4)][0x1F][itemID1(2)][itemID2(2)]...[0x1E]
```

**Separadores:**

- `0x1F` (Unit Separator) - separa campos dentro de um registro
- `0x1E` (Record Separator) - marca fim de cada registro

---

## b) Atributos Multivalorados (String)

Arrays de item IDs são armazenados como:

1. Campo de contagem: `[itemCount(4)]` indicando quantos itens
2. Seguido por IDs consecutivos de 2 bytes: `[itemID1(2)][itemID2(2)]...`

**Localização:** `backend/dao/collection_dao.go`

---

## c) Exclusão Lógica

**Implementação com tombstone bit:**

Cada registro possui um campo tombstone de 1 byte após o ID:

- `0x00` = registro ativo
- `0x01` = registro deletado

**Processo de exclusão:**

1. Localizar registro por ID
2. Verificar se tombstone já é `0x01`
3. Seek para posição do byte tombstone no arquivo
4. Escrever `0x01` para marcar como deletado
5. Sincronizar com disco
6. Incrementar contador de tombstones no header
7. Remover entrada da árvore B+ (apenas para items)

Dados permanecem no arquivo mas são logicamente deletados e ignorados durante leituras.

**Localização:** `backend/dao/item_dao.go` , `backend/dao/collection_dao.go`

---

## d) Chaves Além das PKs

**Apenas Primary Keys (IDs) são indexadas:**

- **Primary Key:** Item ID (inteiro de 2 bytes)
  - Items: IDs auto-incrementais sequenciais
  - Orders: IDs auto-incrementais sequenciais
  - Promotions: IDs auto-incrementais sequenciais

**Foreign Keys (não indexadas):**

- Orders e Promotions armazenam arrays de Item IDs
- Usadas para navegação e referência a items (relacionamento 1:N)
- Sem índices secundários - lookups usam o índice B+ tree de items

---

## e) Estruturas de Dados para Cada Chave

**B+ Tree apenas para Item IDs:**

**Estrutura:**

- Ordem: 4 (máximo 3 chaves por nó)
- Mapeia: `ID (uint64) -> File Offset (int64)`
- Nós folha contêm chaves, offsets e ponteiros para próxima folha (lista ligada)
- Nós internos contêm chaves e ponteiros para filhos
- Todos os dados reais armazenados em nós folha

**Complexidade:**

- Insert: O(log n) com divisão automática de nós
- Search: O(log n) até nós folha
- Delete: Remove entrada da árvore

**Orders e Promotions:** Apenas scan sequencial (sem indexação)

**Localização:** `backend/index/btree.go`, `backend/dao/item_dao.go`

---

## f) Relacionamento 1:N

**Array de foreign keys com integridade referencial:**

**Estrutura:**

- Orders/Promotions armazenam arrays de Item IDs: `ItemIDs []uint64`
- Cada ID ocupa 2 bytes no formato binário

**Navegação:**

- **CreateOrder/CreatePromotion:**
  - Aceita array de item IDs
  - Lê cada item usando `itemDAO.ReadWithIndex()` para validar existência
  - Calcula preço total dos items referenciados
  - Armazena array de IDs no registro order/promotion

**Integridade Referencial:**

- Validada durante criação: se item não existe, criação falha
- Sem cascading deletes: deletar item não remove de orders
- Orders/promotions recuperadas podem referenciar items deletados (tombstoned)

**Padrão:** Collection pattern - OrderDAO e PromotionDAO encapsulam CollectionDAO, que implementa lógica compartilhada para manipular arrays de items.

**Localização:** `app.go` , `backend/dao/collection_dao.go`

---

## g) Persistência dos Índices

**Formato binário em arquivos .idx:**

**Formato:**

```
[count(8)]
[id(8), offset(8)]
[id(8), offset(8)]
...
```

**Processo de salvamento:**

1. Extrai todas entradas da B+ tree usando `GetAll()`
2. Escreve count como inteiro de 8 bytes (big-endian)
3. Escreve cada par (id, offset) como dois inteiros de 8 bytes
4. Chama `file.Sync()` para forçar escrita em disco

**Processo de carregamento:**

1. Lê count
2. Lê cada par (id, offset)
3. Insere em nova B+ tree
4. Retorna árvore populada

**Quando persiste:**

- Após cada insert
- Após cada delete
- Carregado na inicialização do DAO

**Localização:** `backend/index/persistence.go`, arquivo real em `data/items.idx`

---

## h) Estrutura do Projeto no GitHub

**Arquitetura: Wails (Backend Go + Frontend Preact)**

```
BinaryCRUD/
├── backend/
│   ├── dao/              # Camada de Acesso a Dados
│   │   ├── item_dao.go          # Operações CRUD de items
│   │   ├── collection_dao.go    # Lógica compartilhada orders/promotions
│   │   ├── order_dao.go         # Wrapper para orders
│   │   └── promotion_dao.go     # Wrapper para promotions
│   ├── index/            # Implementação B+ Tree
│   │   ├── btree.go             # Estrutura e operações da árvore
│   │   └── persistence.go       # Serialização do índice
│   ├── utils/            # Utilitários binários
│   │   ├── write.go             # Operações de escrita binária
│   │   ├── read.go              # Operações de leitura binária
│   │   ├── header.go            # Gerenciamento de headers
│   │   ├── finder.go            # Busca sequencial
│   │   ├── constants.go         # Separadores ASCII
│   │   └── file.go              # Operações de arquivo
│   └── test/             # Testes unitários
├── frontend/
│   ├── src/              # Componentes Preact
│   │   ├── app.tsx              # Aplicação principal
│   │   └── App.scss             # Estilos
│   └── wailsjs/          # Bindings Wails auto-gerados
├── data/                 # Armazenamento binário
│   ├── items.bin         # Registros de items
│   ├── items.idx         # Índice B+ tree
│   ├── orders.bin        # Registros de orders
│   ├── promotions.bin    # Registros de promotions
│   └── items.json        # Dados de exemplo
├── logs/                 # Logs da aplicação
│   └── app.log           # Log persistente
├── app.go                # API backend (bindings Wails)
├── main.go               # Ponto de entrada da aplicação
├── logger.go             # Sistema de logging (slog)
└── README.md             # Documentação

```

**Padrões de Design:**

- **DAO Pattern:** Separação de lógica de acesso a dados
- **Wrapper Pattern:** OrderDAO/PromotionDAO encapsulam CollectionDAO
- **Mutex Protection:** Operações concorrentes thread-safe
- **Binary Format:** Serialização customizada com tombstones e separadores
- **Handler Pattern:** InMemoryHandler implementa interface slog.Handler
