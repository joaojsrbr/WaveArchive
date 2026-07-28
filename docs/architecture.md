# Arquitetura inicial

## Decisão

A primeira entrega é uma fatia vertical do catálogo de personagens:

```text
Nanoka → DTO externo → mapper → domínio → caso de uso → SQLite → binding Wails → React
```

O domínio não importa Wails, SQLite ou estruturas do Nanoka. O aplicativo Wails
apenas compõe as dependências e expõe casos de uso.

## O que foi reaproveitado do scraper

- Detecção de `ww.latest` no manifest e remoção do sufixo `+hash`.
- Endpoint do índice de personagens e timeout de 20 segundos.
- Mapeamentos observados de elemento e tipo de arma.
- Regras de `cleanText` e `applyParams`, incluindo parâmetros aninhados,
  tags de plataforma/gênero e preservação de placeholders desconhecidos.
- Paths de ícone e background são preservados para a fase de cache de assets.
- Arma assinatura, detalhes, guias e exportadores permanecem no roadmap imediato.

O arquivo `wuwa_scraper.py` é uma referência somente leitura e não é usado em
produção.

## Persistência

Os dados sincronizados ficam em `characters`, `game_versions`, `materials`,
`character_ascension_costs`, `character_level_exp`, `character_stats`,
`skill_progression`, `skill_unlock_costs` e `skill_level_costs`. O catálogo
rico `item_all.json` fornece metadados e fontes dos materiais; o índice
`item.json` é mesclado como fallback. Dados pessoais
ficam em `owned_characters`; uma sincronização faz upsert somente nos dados
oficiais e não substitui posse, sequência ou favorito.

O planejador de ascensão recebe nível atual/alvo e níveis atuais/alvo de cada
habilidade. A seleção e agregação dos custos é uma regra do caso de uso Go; o
React apenas coleta escolhas e apresenta ascensão, habilidades e total. Os
multiplicadores oficiais permanecem normalizados como linhas por habilidade e
são apresentados na aba Kit.

A ordem original das chaves de `character.json` é capturada durante o decode,
antes da conversão para `map`, e persistida em `characters.api_order`. Isso
permite reproduzir offline a ordenação editorial/de lançamento fornecida pela
API sem tentar deduzi-la a partir do ID do personagem.

Tags de função, histórias, objetos pessoais, guia do Forte, parâmetros de
fraqueza, dependências da árvore e ramificações especiais são convertidos para
`CharacterExtras` no domínio. A persistência usa `characters.extras_json`
porque esses documentos são lidos como um agregado do personagem e o formato
da fonte varia entre versões; o React recebe tipos próprios e não conhece os
DTOs da Nanoka.

Os totais de habilidade seguem a semântica observada no planejador oficial:
níveis 2–10 das cinco habilidades principais, custo de nível 1 das habilidades
inerentes e `consume` dos nós de atributo. Custos genéricos de Outro/Tune Break
não entram. Materiais agregados são exibidos pela ordem numérica da API, com
Shell Credits primeiro.

O banco é criado em `%AppData%\WaveArchive\wavearchive.db`, usa WAL,
foreign keys, busy timeout e migrations incorporadas ao executável. Antes de
uma atualização de um catálogo existente, `VACUUM INTO` cria um snapshot
consistente. A restauração cria antes uma cópia de segurança do estado atual.

Imagens são baixadas como WebP para `%AppData%\WaveArchive\assets` com limite
de tamanho, validação do Content-Type e rename atômico. O asset server do Wails
expõe somente o prefixo `/cache/`; nenhum servidor HTTP adicional é criado.

## Próximas decisões

- Expandir o controle de inventário para descontar materiais já possuídos.
- Implementar staging/snapshot antes de ampliar a sincronização.
- Gerar thumbnails e adicionar política configurável de limite/expiração ao
  cache local já implementado.
- Adicionar perfis locais antes de habilitar edição de posse.
