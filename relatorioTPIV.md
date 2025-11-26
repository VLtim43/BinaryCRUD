# Relatório Técnico - TP IV

## Formulário Técnico

### 1. Qual foi a taxa de compressão obtida com o algoritmo de Huffman?

**Arquivo testado:** `items.bin`

**a. Tamanho do arquivo original:** 1011 bytes

**b. Tamanho do arquivo comprimido:** 954 bytes

**c. Cálculo da taxa:**
```
Taxa = (tamanho_comprimido / tamanho_original) × 100
Taxa = (954 / 1011) × 100 = 94.36%
Economia = 100% - 94.36% = 5.64%
```

**d. Interpretação do resultado:**

O algoritmo Huffman obteve uma economia de 5.64% do tamanho original. A compressão relativamente baixa ocorre porque arquivos binários estruturados (como `.bin`) já possuem dados compactos com distribuição mais uniforme de bytes. Além disso, o overhead do header (árvore serializada + metadados) impacta proporcionalmente mais em arquivos pequenos. Em arquivos maiores ou com mais redundância textual, a taxa seria superior.

---

### 2. Qual foi a taxa de compressão obtida com o algoritmo de LZW?

**Arquivo testado:** `items.bin`

**a. Tamanho do arquivo original:** 1011 bytes

**b. Tamanho do arquivo comprimido:** 1202 bytes

**c. Cálculo da taxa:**
```
Taxa = (tamanho_comprimido / tamanho_original) × 100
Taxa = (1202 / 1011) × 100 = 118.89%
Economia = 100% - 118.89% = -18.89% (expansão)
```

**d. Interpretação do resultado:**

O LZW apresentou expansão de 18.89% ao invés de compressão. Isso ocorre porque a implementação utiliza códigos de 16 bits fixos para cada entrada do dicionário. Em arquivos pequenos (1011 bytes) com poucos padrões repetitivos, cada byte de entrada gera um código de 2 bytes na saída, resultando em expansão. Além disso, o header do formato (magic bytes + metadados) adiciona overhead fixo. O LZW é mais eficiente em arquivos maiores com muitas sequências repetidas, onde o dicionário consegue representar padrões longos com códigos curtos.

---

### 3. Quais dificuldades surgiram ao implementar Huffman e LZW e como você resolveu?

**Huffman:**

1. **Serialização da árvore:** A árvore precisa ser salva junto com os dados comprimidos para descompressão. Solução: formato `[1][byte]` para folhas e `[0][left][right]` para nós internos, permitindo reconstrução recursiva.

2. **Padding de bits:** O último byte pode ter bits não utilizados. Solução: armazenar quantidade de bits de padding no header e removê-los na descompressão.

3. **Caso especial de único símbolo:** Arquivo com um único byte repetido não gera árvore válida. Solução: criar nó artificial com código "0" para esse caso.

**LZW:**

1. **Caso especial KwKwK:** Quando o código ainda não existe no dicionário durante descompressão. Solução: detectar quando `code == nextCode` e construir a entrada como `current + current[0]`.

2. **Aliasing de slices:** Modificações em entradas do dicionário afetavam outras. Solução: fazer cópias explícitas com `copy()` ao adicionar ao dicionário.

3. **Overflow do dicionário:** Dicionário pode crescer indefinidamente. Solução: limitar a 65535 entradas (máximo de uint16).

---

### 4. Justifique a escolha da estrutura de dados usada para armazenar as tabelas, dicionários e árvores utilizados pelos algoritmos.

**Huffman - Árvore Binária:**

```go
type HuffmanNode struct {
    Byte   byte
    Freq   int
    Left   *HuffmanNode
    Right  *HuffmanNode
    IsLeaf bool
}
```

Estrutura de árvore binária com ponteiros é ideal para Huffman pois:
- Permite travessia natural pelos códigos (0=esquerda, 1=direita)
- Facilita serialização recursiva
- Construção bottom-up com heap de prioridade

**Huffman - Min-Heap (Priority Queue):**

```go
type nodeHeap []*HuffmanNode
```

Heap de prioridade para construir a árvore, garantindo que os dois nós de menor frequência sejam sempre combinados primeiro - requisito fundamental do algoritmo.

**Huffman - Mapa de Códigos:**

```go
codeMap map[byte]string
```

HashMap byte→string para lookup O(1) durante compressão. A string representa a sequência de bits como "0110".

**LZW - Dicionário de Compressão:**

```go
dictionary map[string]uint16
```

HashMap string→código para verificação rápida se sequência existe no dicionário. Chave é a sequência de bytes, valor é o código atribuído.

**LZW - Dicionário de Descompressão:**

```go
dictionary map[uint16][]byte
```

HashMap inverso código→bytes para reconstruir as sequências originais durante descompressão.

---

### 5. Qual campo foi escolhido para criptografia? Por quê?

**Campo:** `ownerOrName` (nome do cliente em Orders / nome da promoção em Promotions)

**Justificativa:**

1. **Dado sensível:** Nome de clientes é informação pessoal que requer proteção (LGPD)
2. **Presente em múltiplas tabelas:** Demonstra criptografia funcionando em Orders e Promotions
3. **Tamanho variável:** Testa a criptografia com strings de diferentes tamanhos
4. **Verificação visual:** Fácil confirmar que está criptografado olhando o arquivo binário
5. **Não é chave primária:** Evita complicações com índices (IDs permanecem em plaintext para busca)

---

### 6. Descreva como o RSA foi implementado no projeto.

**a. Estrutura das chaves pública e privada:**

```go
type RSACrypto struct {
    privateKey *rsa.PrivateKey  // Contém também a chave pública
    publicKey  *rsa.PublicKey
}
```

A chave privada RSA do Go (`rsa.PrivateKey`) contém os componentes:
- `N`: módulo (produto de dois primos)
- `E`: expoente público (geralmente 65537)
- `D`: expoente privado
- `Primes`: os fatores primos p e q
- `Precomputed`: valores pré-calculados para otimização

**b. Como e onde foram armazenadas:**

As chaves são armazenadas em `data/keys/` no formato PEM:

- `private.pem` - Chave privada (permissão 0600)
- `public.pem` - Chave pública (permissão 0644)

Formato PEM com headers `-----BEGIN RSA PRIVATE KEY-----` e `-----BEGIN RSA PUBLIC KEY-----`.

**c. Como foram carregadas pelo sistema:**

```go
func GetInstance() (*RSACrypto, error) {
    once.Do(func() {
        instance = &RSACrypto{}
        initErr = instance.loadOrGenerateKeys()
    })
    return instance, nil
}
```

Padrão Singleton com `sync.Once`:
1. Na primeira chamada, verifica se `private.pem` existe
2. Se existe: carrega com `x509.ParsePKCS1PrivateKey()`
3. Se não existe: gera novo par com `rsa.GenerateKey()` e salva

**d. Tamanho das chaves escolhidas e justificativa:**

**Tamanho:** 2048 bits

**Justificativa:**
- Considerado seguro até ~2030 segundo NIST
- Permite criptografar até ~190 bytes por operação (suficiente para nomes)
- Balanceia segurança vs performance
- 1024 bits é considerado fraco; 4096 bits seria excessivo para este caso de uso

**e. Em qual momento a criptografia do(s) campo(s) ocorre (no CRUD):**

**Create (Write):** `collection_dao.go:55-63`

```go
func (dao *CollectionDAO) Write(ownerOrName string, ...) (uint64, error) {
    // Encrypt ANTES de gravar
    rsaCrypto, _ := crypto.GetInstance()
    encryptedName, _ := rsaCrypto.EncryptString(ownerOrName)

    // encryptedName vai para o arquivo .bin
    nameBytes := encryptedName
    // ... AppendEntry grava os bytes criptografados
}
```

A criptografia ocorre na camada DAO, antes de qualquer escrita no arquivo binário.

**f. Em qual momento ocorre a descriptografia:**

**Read:** `collection_dao.go:190-198`

```go
func (dao *CollectionDAO) readUnlocked(id uint64) (*Collection, error) {
    // Lê bytes criptografados do arquivo
    collection, _ := utils.ParseCollectionEntry(entryData)

    // Decrypt DEPOIS de ler
    rsaCrypto, _ := crypto.GetInstance()
    decryptedName, _ := rsaCrypto.DecryptString([]byte(collection.OwnerOrName))

    return &Collection{
        OwnerOrName: decryptedName,  // Retorna plaintext
        // ...
    }
}
```

**GetAll:** `collection_dao.go:257-262` - Descriptografa cada registro no loop.

**g. Conversões realizadas:**

```
CRIPTOGRAFIA (Write):
string "John Doe"
    ↓ []byte(plaintext)
[]byte{74, 111, 104, 110, 32, 68, 111, 101}
    ↓ rsa.EncryptOAEP(sha256, publicKey, bytes)
[]byte{encrypted...} // 256 bytes (tamanho fixo RSA-2048)
    ↓ grava no .bin
[arquivo binário com 256 bytes criptografados]

DESCRIPTOGRAFIA (Read):
[arquivo binário]
    ↓ lê bytes
[]byte{encrypted...} // 256 bytes
    ↓ rsa.DecryptOAEP(sha256, privateKey, bytes)
[]byte{74, 111, 104, 110, 32, 68, 111, 101}
    ↓ string(plaintext)
"John Doe"
```

O RSA-OAEP com SHA-256 adiciona padding aleatório, então:
- Input de qualquer tamanho (até ~190 bytes) → Output fixo de 256 bytes
- Mesmo plaintext gera ciphertexts diferentes (devido ao padding aleatório)

---

## Estrutura de Diretórios Atualizada

```
BinaryCRUD/
├── backend/
│   ├── compression/
│   │   ├── huffman.go      # Compressão Huffman
│   │   └── lzw.go          # Compressão LZW
│   ├── crypto/
│   │   └── rsa.go          # Criptografia RSA
│   ├── dao/
│   │   ├── collection_dao.go  # DAO com encrypt/decrypt
│   │   └── ...
│   └── ...
├── data/
│   ├── bin/                # Arquivos .bin (dados criptografados)
│   ├── compressed/         # Arquivos comprimidos
│   ├── indexes/            # Índices B+ Tree
│   ├── keys/               # Chaves RSA (private.pem, public.pem)
│   └── seed/               # Dados iniciais JSON
└── ...
```
