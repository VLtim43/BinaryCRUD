# Relatório Técnico - TP V

## Formulário Técnico

### 1. Qual campo textual foi escolhido para aplicar os algoritmos de casamento de padrões? Por quê?

**Campo:** `name` (nome do item na tabela Items)

**Justificativa:**

1. **Campo textual significativo:** O nome do item é o principal campo de texto pesquisável na aplicação
2. **Variedade de padrões:** Nomes como "Classic Burger", "Chocolate Milkshake" permitem testar padrões variados
3. **Caso de uso real:** Usuários frequentemente buscam produtos por nome parcial (ex: "burger", "chicken")
4. **Tamanho adequado:** Nomes têm tamanho suficiente para demonstrar a eficiência dos algoritmos
5. **Não criptografado:** Diferente do campo `ownerOrName` em Orders (que usa RSA), o nome do item está em plaintext, facilitando a busca

---

### 2. Explique o funcionamento do KMP implementado

**Arquivo:** `backend/search/kmp.go`

**a. Estrutura:**

```go
type KMP struct {
    pattern []byte
    lps     []int  // Longest Proper Prefix which is also Suffix
}
```

**b. Pré-processamento - Construção da tabela LPS:**

```go
func computeLPS(pattern []byte) []int {
    lps := make([]int, len(pattern))
    length := 0
    i := 1

    for i < len(pattern) {
        if pattern[i] == pattern[length] {
            length++
            lps[i] = length
            i++
        } else {
            if length != 0 {
                length = lps[length-1]  // Usa valor anterior
            } else {
                lps[i] = 0
                i++
            }
        }
    }
    return lps
}
```

A tabela LPS armazena, para cada posição `i`, o comprimento do maior prefixo próprio que também é sufixo do padrão `pattern[0..i]`.

**Exemplo:** Para o padrão "AAACAAAA":
```
Padrão: A  A  A  C  A  A  A  A
LPS:    0  1  2  0  1  2  3  3
```

**c. Busca:**

```go
func (k *KMP) Search(text []byte) []int {
    i := 0  // índice no texto
    j := 0  // índice no padrão

    for i < len(text) {
        if pattern[j] == text[i] {
            i++
            j++
        }

        if j == len(pattern) {
            // Match encontrado na posição i-j
            matches = append(matches, i-j)
            j = lps[j-1]  // Continua buscando sobreposições
        } else if i < len(text) && pattern[j] != text[i] {
            if j != 0 {
                j = lps[j-1]  // Pula usando LPS (não volta no texto!)
            } else {
                i++
            }
        }
    }
}
```

**Complexidade:** O(n + m), onde n = tamanho do texto, m = tamanho do padrão.

---

### 3. Explique o funcionamento do Boyer–Moore implementado

**Arquivo:** `backend/search/boyer_moore.go`

**a. Estrutura:**

```go
type BoyerMoore struct {
    pattern      []byte
    badCharTable [256]int  // Tabela de deslocamento Bad Character
}
```

**b. Pré-processamento - Tabela Bad Character:**

```go
func (bm *BoyerMoore) computeBadCharTable() {
    m := len(bm.pattern)

    // Inicializa com deslocamento máximo
    for i := 0; i < 256; i++ {
        bm.badCharTable[i] = m
    }

    // Para caracteres no padrão: distância até o final
    for i := 0; i < m-1; i++ {
        bm.badCharTable[bm.pattern[i]] = m - 1 - i
    }
}
```

**Exemplo:** Para o padrão "EXEMPLO":
```
E: 7-1-0 = 6  (posição 0)
X: 7-1-1 = 5  (posição 1)
M: 7-1-3 = 3  (posição 3)
P: 7-1-4 = 2  (posição 4)
L: 7-1-5 = 1  (posição 5)
O: última posição, não entra
Outros: 7 (tamanho do padrão)
```

**c. Busca (da direita para esquerda):**

```go
func (bm *BoyerMoore) Search(text []byte) []int {
    i := 0  // posição no texto onde o padrão começa

    for i <= len(text)-len(pattern) {
        j := len(pattern) - 1  // Começa do FINAL do padrão

        // Compara da direita para esquerda
        for j >= 0 && pattern[j] == text[i+j] {
            j--
        }

        if j < 0 {
            // Match encontrado!
            matches = append(matches, i)
            i++
        } else {
            // Mismatch: usa heurística Bad Character
            shift := badCharTable[text[i+j]] - (m - 1 - j)
            if shift < 1 {
                shift = 1
            }
            i += shift
        }
    }
}
```

**Complexidade:** O(n/m) melhor caso, O(n*m) pior caso.

**Vantagem:** Em textos grandes com alfabetos grandes, Boyer-Moore frequentemente pula grandes porções do texto sem examinar cada caractere.

---

### 4. Descreva como integrou os algoritmos ao sistema

**a. Camada de Busca (`backend/search/`):**

Criados dois arquivos independentes:
- `kmp.go` - Implementação KMP
- `boyer_moore.go` - Implementação Boyer-Moore

Ambos expõem interface similar:
```go
type Matcher interface {
    ContainsString(text string) bool
}
```

**b. Camada DAO (`backend/dao/item_dao.go`):**

```go
type SearchAlgorithm string

const (
    AlgorithmKMP        SearchAlgorithm = "kmp"
    AlgorithmBoyerMoore SearchAlgorithm = "bm"
)

func (dao *ItemDAO) SearchByName(pattern string, algorithm SearchAlgorithm) ([]Item, error) {
    items, _ := dao.GetAll()
    lowerPattern := strings.ToLower(pattern)

    // Seleciona algoritmo
    var matcher interface{ ContainsString(string) bool }
    switch algorithm {
    case AlgorithmBoyerMoore:
        matcher = search.NewBoyerMooreString(lowerPattern)
    default:
        matcher = search.NewKMPString(lowerPattern)
    }

    // Busca case-insensitive
    for _, item := range items {
        if matcher.ContainsString(strings.ToLower(item.Name)) {
            results = append(results, item)
        }
    }
    return results, nil
}
```

**c. Camada API (`app.go`):**

```go
func (a *App) SearchItems(pattern string, algorithm string) ([]map[string]any, error) {
    // Converte string para enum
    var searchAlgo dao.SearchAlgorithm
    switch algorithm {
    case "bm":
        searchAlgo = dao.AlgorithmBoyerMoore
    default:
        searchAlgo = dao.AlgorithmKMP
    }

    items, _ := a.itemDAO.SearchByName(pattern, searchAlgo)
    // ... retorna resultados
}
```

**d. Frontend (`frontend/src/components/tabs/ItemTab.tsx`):**

```tsx
<Select
    value={searchAlgorithm}
    options={[
        { value: "kmp", label: "KMP (Knuth-Morris-Pratt)" },
        { value: "bm", label: "Boyer-Moore" },
    ]}
/>
<Input placeholder="Search by name..." value={searchQuery} />
<Button onClick={handleSearch}>Search</Button>
```

**Fluxo completo:**
```
Frontend → SearchItems(pattern, "kmp"|"bm")
         → ItemDAO.SearchByName(pattern, algorithm)
         → KMP.ContainsString() ou BoyerMoore.ContainsString()
         → Retorna itens encontrados
```

---

### 5. Quais dificuldades encontrou na implementação dos dois algoritmos?

**KMP:**

1. **Construção correta da tabela LPS:** O caso de voltar usando `lps[length-1]` quando há mismatch foi confuso inicialmente. Solução: seguir o pseudocódigo clássico e testar com padrões conhecidos como "AAACAAAA".

2. **Busca de sobreposições:** Após encontrar um match, usar `j = lps[j-1]` para continuar buscando padrões sobrepostos (ex: "aa" em "aaaa" deve encontrar posições 0, 1, 2).

**Boyer-Moore:**

1. **Cálculo do deslocamento:** Entender que o shift é `badChar[c] - (m - 1 - j)` e não simplesmente `badChar[c]`. O ajuste considera a posição atual no padrão.

2. **Shift mínimo de 1:** Quando o cálculo resulta em shift ≤ 0 (caractere encontrado à direita da posição atual), forçar shift = 1 para garantir progresso.

3. **Última posição do padrão:** Não incluir a última posição no cálculo da bad char table para evitar shifts de 0.

**Integração:**

1. **Case-insensitive:** Converter tanto o padrão quanto cada nome para lowercase antes de comparar, mantendo os dados originais intactos.

2. **Interface comum:** Criar interface `ContainsString(string) bool` permitiu usar ambos algoritmos de forma intercambiável no DAO.

---

## Estrutura de Arquivos Adicionados

```
BinaryCRUD/
├── backend/
│   ├── search/
│   │   ├── kmp.go           # Algoritmo KMP
│   │   └── boyer_moore.go   # Algoritmo Boyer-Moore
│   ├── test/
│   │   ├── kmp_test.go      # Testes KMP
│   │   └── boyer_moore_test.go  # Testes Boyer-Moore
│   └── dao/
│       └── item_dao.go      # SearchByName com seleção de algoritmo
└── frontend/
    └── src/
        └── components/
            └── tabs/
                └── ItemTab.tsx  # UI com seletor de algoritmo
```
