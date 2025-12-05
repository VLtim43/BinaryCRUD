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

A estrutura `KMP` contém o padrão como array de bytes e a tabela LPS (Longest Proper Prefix which is also Suffix) como array de inteiros.

**b. Pré-processamento - Construção da tabela LPS:**

A tabela LPS armazena, para cada posição `i`, o comprimento do maior prefixo próprio que também é sufixo do padrão `pattern[0..i]`. O algoritmo percorre o padrão uma vez, comparando caracteres e usando valores já calculados para evitar recomputação.

**Exemplo:** Para o padrão "AAACAAAA":

| Posição | A | A | A | C | A | A | A | A |
|---------|---|---|---|---|---|---|---|---|
| LPS     | 0 | 1 | 2 | 0 | 1 | 2 | 3 | 3 |

**c. Busca:**

O algoritmo mantém dois índices: `i` para o texto e `j` para o padrão. Quando há match, ambos avançam. Quando há mismatch:

- Se `j > 0`: usa a tabela LPS para pular (`j = lps[j-1]`), sem retroceder no texto
- Se `j == 0`: avança `i` no texto

Ao encontrar match completo (`j == len(pattern)`), registra a posição e usa LPS para continuar buscando sobreposições.

**Complexidade:** O(n + m), onde n = tamanho do texto, m = tamanho do padrão.

---

### 3. Explique o funcionamento do Boyer-Moore implementado

**Arquivo:** `backend/search/boyer_moore.go`

**a. Estrutura:**

A estrutura `BoyerMoore` contém o padrão como array de bytes e a tabela Bad Character como array de 256 inteiros (um para cada valor de byte possível).

**b. Pré-processamento - Tabela Bad Character:**

Para cada caractere do alfabeto, armazena a distância da sua última ocorrência no padrão até o final. Caracteres ausentes no padrão recebem o tamanho total do padrão como valor.

**Exemplo:** Para o padrão "EXEMPLO" (tamanho 7):

| Caractere | E | X | M | P | L | Outros |
|-----------|---|---|---|---|---|--------|
| Shift     | 6 | 5 | 3 | 2 | 1 | 7      |

**c. Busca (da direita para esquerda):**

Diferente do KMP, Boyer-Moore compara o padrão da direita para esquerda. Quando há mismatch:

1. Consulta a tabela Bad Character para o caractere do texto que causou o mismatch
2. Calcula o deslocamento: `shift = badChar[c] - (m - 1 - j)`
3. Se shift ≤ 0, força shift = 1 para garantir progresso
4. Desloca o padrão pelo valor calculado

**Complexidade:** O(n/m) melhor caso, O(n*m) pior caso.

**Vantagem:** Em textos grandes com alfabetos grandes, Boyer-Moore frequentemente pula grandes porções do texto sem examinar cada caractere.

---

### 4. Descreva como integrou os algoritmos ao sistema

**a. Camada de Busca (`backend/search/`):**

Criados dois arquivos independentes:

- `kmp.go` - Implementação KMP
- `boyer_moore.go` - Implementação Boyer-Moore

Ambos expõem interface similar com métodos `Search()` (retorna todas as posições) e `ContainsString()` (retorna boolean).

**b. Camada DAO (`backend/dao/item_dao.go`):**

O método `SearchByName` recebe o padrão e o algoritmo desejado ("kmp" ou "bm"). Internamente:

1. Obtém todos os itens do banco
2. Converte o padrão para lowercase
3. Instancia o matcher apropriado (KMP ou Boyer-Moore)
4. Para cada item, verifica se o nome (em lowercase) contém o padrão
5. Retorna os itens que casaram

**c. Camada API (`app.go`):**

A função `SearchItems` recebe o padrão e nome do algoritmo como strings, converte para o tipo interno e chama o DAO.

**d. Frontend (`frontend/src/components/tabs/ItemTab.tsx`):**

Interface com:

- Select para escolher algoritmo (KMP ou Boyer-Moore)
- Input para digitar o padrão de busca
- Botão para executar a busca
- Exibição dos resultados

**Fluxo completo:**

Frontend → SearchItems(pattern, "kmp"|"bm") → ItemDAO.SearchByName() → KMP ou BoyerMoore → Retorna itens encontrados

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

2. **Interface comum:** Criar método `ContainsString(string) bool` em ambos permitiu usá-los de forma intercambiável no DAO.

---

## Estrutura de Arquivos Adicionados

```text
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
