# BinaryCRUD

**BinaryCRUD** is a restaurant manager application built with Wails and Preact-TypeScript that implements custom binary file-based CRUD operations. It manages items (products) and orders using binary serialization with B+ Tree indexing, logical deletion with tombstones, and 1:N relationships stored inline.

## To compile it locally

- Install Go <https://go.dev/doc/install>
- Install Wails <https://wails.io/docs/gettingstarted/installation/>
- May need to install some libs like libwebkit. Perhaps node and npm too
- run with ./run.sh

## Relatório Técnico

## a) Estrutura de Registro

- `backend/persistence/item.go`: struct `Item` (`RecordID`, `Name`, `Tombstone`, `Timestamp`) representa cada produto; o `RecordID` é o identificador lógico.
- `backend/persistence/order.go`: structs `Order` (cabeçalho de pedido com `Items`, `Tombstone`, `Timestamp`) e `OrderItem` (par `ItemID`/`Quantity`) modelam o relacionamento entre pedidos e itens.

## b) Atributos multivalorados de string

- O único campo textual persistido é `Item.Name`. Cada nome é gravado como `[NameLength:uint32][NameBytes]`, ladeado por separadores 0x1F para preservar integridade e permitir comprimentos variáveis (`backend/persistence/item.go`).

## c) Exclusão lógica

- `ItemDAO.Delete` usa o índice para recuperar o deslocamento do registro, abre o arquivo em modo leitura/escrita e sobrescreve o byte de tombstone com `1`, marcando o item como excluído sem remover seu espaço (`backend/dao/item_dao.go`).
- Leitores (`ReadItemRecord`) convertem o byte para booleano e filtram registros tombstoned nas consultas (`backend/persistence/item.go`, `backend/persistence/read.go`).

## d) Chaves utilizadas

- A chave primária lógica é o `RecordID` sequencial atribuído no append.

## e) Estruturas de índice

- Itens (`data/items.bin`) usam uma B+ Tree (`backend/index/b_tree`) para mapear `RecordID → offset`.
- Pedidos não possuem índice adicional; `OrderDAO.GetByID` realiza varredura sequencial sobre `orders.bin`.

## f) Relacionamento 1:N

- O relacionamento `Order → OrderItems` é serializado inline no registro do pedido: `[ItemCount:uint32]` seguido por pares `[ItemID:uint32][Quantity:uint32]`, cada qual separado por 0x1F (`backend/persistence/order.go`).
- A navegação 1:N lê o pedido completo, iterando sobre o array `Items`. Não há enforcement automático de integridade referencial; cabe à camada de aplicação garantir que `ItemID` se refira a itens existentes (`app.go`).

## g) Persistência dos índices

- O índice B+ Tree é persistido em `<arquivo>.idx` com formato binário fixo: cabeçalho (`"BIDX"`, versão, ordem, contagem) seguido por pares `[Key:uint32][Offset:int64]`. O salvamento ocorre após inserções (`backend/index/b_tree/serialize.go`, `backend/index/index_manager.go`).
- Na inicialização, `IndexManager.Initialize` tenta carregar o arquivo; em caso de falha, reconstrói varrendo os registros e sincroniza novamente com disco.

## h) Organização do projeto

- `main.go` e `app.go`: inicialização Wails e camada de aplicação/serviços exposta ao frontend.
- `backend/dao`: Data Access Objects (itens e pedidos), coordenação entre persistência e indexação.
- `backend/persistence`: leitura/escrita binária comum, layouts de registros, append, leitura, impressão.
- `backend/index`: implementação da B+ Tree e gerenciador de índices.
- `frontend/`: aplicação web (Vite/TypeScript) empacotada pelo Wails; build final em `frontend/dist`.
