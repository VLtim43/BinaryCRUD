# Relatório Técnico - TP IV

## Formulário Técnico

### 1. Qual foi a taxa de compressão obtida com o algoritmo de Huffman?

**Arquivo testado:** `items.bin`

**a. Tamanho do arquivo original:** 1011 bytes

**b. Tamanho do arquivo comprimido:** 954 bytes

**c. Cálculo da taxa:**

Taxa = (954 / 1011) × 100 = 94.36%
Economia = 100% - 94.36% = 5.64%

**d. Interpretação do resultado:**

O algoritmo Huffman obteve uma economia de 5.64% do tamanho original. A compressão relativamente baixa ocorre porque arquivos binários estruturados (como `.bin`) já possuem dados compactos com distribuição mais uniforme de bytes. Além disso, o overhead do header (árvore serializada + metadados) impacta proporcionalmente mais em arquivos pequenos. Em arquivos maiores ou com mais redundância textual, a taxa seria superior.

---

### 2. Qual foi a taxa de compressão obtida com o algoritmo de LZW?

**Arquivo testado:** `items.bin`

**a. Tamanho do arquivo original:** 1011 bytes

**b. Tamanho do arquivo comprimido:** 1202 bytes

**c. Cálculo da taxa:**

Taxa = (1202 / 1011) × 100 = 118.89%
Economia = 100% - 118.89% = -18.89% (expansão)

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

2. **Aliasing de slices:** Modificações em entradas do dicionário afetavam outras. Solução: fazer cópias explícitas ao adicionar ao dicionário.

3. **Overflow do dicionário:** Dicionário pode crescer indefinidamente. Solução: limitar a 65535 entradas (máximo de uint16).

---

### 4. Justifique a escolha da estrutura de dados usada para armazenar as tabelas, dicionários e árvores utilizados pelos algoritmos.

**Huffman - Árvore Binária:**

Estrutura `HuffmanNode` com ponteiros para filhos esquerdo e direito, byte armazenado, frequência e flag indicando se é folha. Estrutura de árvore binária com ponteiros é ideal para Huffman pois:
- Permite travessia natural pelos códigos (0=esquerda, 1=direita)
- Facilita serialização recursiva
- Construção bottom-up com heap de prioridade

**Huffman - Min-Heap (Priority Queue):**

Slice de ponteiros para `HuffmanNode` implementando a interface `heap.Interface`. Heap de prioridade para construir a árvore, garantindo que os dois nós de menor frequência sejam sempre combinados primeiro - requisito fundamental do algoritmo.

**Huffman - Mapa de Códigos:**

HashMap `byte → string` para lookup O(1) durante compressão. A string representa a sequência de bits como "0110".

**LZW - Dicionário de Compressão:**

HashMap `string → uint16` para verificação rápida se sequência existe no dicionário. Chave é a sequência de bytes, valor é o código atribuído.

**LZW - Dicionário de Descompressão:**

HashMap inverso `uint16 → []byte` para reconstruir as sequências originais durante descompressão.

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

**Nota:** Esta é uma implementação educacional do RSA para fins didáticos, demonstrando os conceitos matemáticos do algoritmo.

**a. Estrutura das chaves pública e privada:**

A estrutura `SimpleRSA` contém:
- `n`: módulo (produto de dois primos p × q)
- `e`: expoente público (padrão 65537, com fallback para valores menores)
- `d`: expoente privado (inverso modular de e mod φ(n))
- `phi`: φ(n) = (p-1)(q-1) - totiente de Euler

**b. Geração das chaves:**

As chaves são geradas em tempo de execução com primos configuráveis (padrão: p=61, q=53):
1. Calcula n = p × q (módulo)
2. Calcula φ(n) = (p-1)(q-1)
3. Escolhe e tal que gcd(e, φ(n)) = 1
4. Calcula d usando o algoritmo de Euclides estendido

**c. Como foram carregadas pelo sistema:**

Padrão Singleton com `sync.Once` garante que apenas uma instância seja criada. A instância é inicializada na primeira chamada a `GetInstance()` e reutilizada em todas as operações subsequentes.

**d. Tamanho das chaves e justificativa:**

**Tamanho:** Primos pequenos (p=61, q=53, resultando em n=3233)

**Justificativa educacional:**
- Permite demonstrar claramente a matemática do RSA
- Valores pequenos facilitam debug e verificação manual
- Foco no entendimento do algoritmo, não em segurança de produção

**e. Em qual momento a criptografia do(s) campo(s) ocorre (no CRUD):**

**Create (Write):** No `collection_dao.go`, antes de gravar no arquivo binário:
1. Recebe o nome em plaintext
2. Chama `rsaCrypto.EncryptToBytes(ownerOrName)`
3. Grava os bytes criptografados no arquivo `.bin`

A criptografia ocorre na camada DAO, antes de qualquer escrita no arquivo binário.

**f. Em qual momento ocorre a descriptografia:**

**Read:** No `collection_dao.go`, após ler do arquivo binário:
1. Lê os bytes criptografados do arquivo
2. Chama `rsaCrypto.DecryptFromBytes(encryptedBytes)`
3. Retorna o nome descriptografado para a aplicação

**GetAll:** Descriptografa cada registro individualmente no loop de leitura.

**g. Conversões realizadas:**

**Criptografia (Write):**
1. String plaintext → array de bytes
2. Para cada byte: c = m^e mod n (exponenciação modular)
3. Serializa os big.Int resultantes com prefixo de tamanho
4. Grava no arquivo binário

**Descriptografia (Read):**
1. Lê bytes do arquivo binário
2. Desserializa para array de big.Int
3. Para cada valor: m = c^d mod n (exponenciação modular)
4. Converte bytes de volta para string

A exponenciação modular usa o algoritmo square-and-multiply para eficiência.

---

## Estrutura de Diretórios Atualizada

```
BinaryCRUD/
├── backend/
│   ├── compression/
│   │   ├── huffman.go      # Compressão Huffman
│   │   └── lzw.go          # Compressão LZW
│   ├── crypto/
│   │   └── simple_rsa.go   # Criptografia RSA (educacional)
│   ├── dao/
│   │   ├── collection_dao.go  # DAO com encrypt/decrypt
│   │   └── ...
│   └── ...
├── data/
│   ├── bin/                # Arquivos .bin (dados criptografados)
│   ├── compressed/         # Arquivos comprimidos
│   ├── indexes/            # Índices B+ Tree
│   └── seed/               # Dados iniciais JSON
└── ...
```
