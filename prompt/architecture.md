# Arquitetura viva — WaveArchive

Última revisão: 2026-07-29.

## Produto

WaveArchive é um aplicativo desktop local-first para consulta, planejamento e
gerenciamento de dados de Wuthering Waves. O executável não depende do
`wuwa_scraper.py`; esse arquivo é somente uma referência funcional.

## Stack

- Backend: Go 1.23.
- Desktop: Wails v2.
- Frontend: React 19, TypeScript e Vite.
- Persistência: SQLite via `modernc.org/sqlite`.
- Interface: tema escuro premium com ícones Lucide.
- Fontes externas: Nanoka, catálogo de branches do Arikatsu Data e servidor de
  guias configurado no código.
- IA: Ollama, LM Studio e Gemini, atrás do analisador Go.

## Fluxo principal

```text
Nanoka / Arikatsu Data / guias
      ↓
clientes em internal/sources
      ↓
DTO externo e mapper
      ↓
tipos em internal/domain
      ↓
casos de uso em internal/usecase
      ↓
repositórios SQLite em internal/repository
      ↓
métodos expostos por app.go / Wails
      ↓
frontend/src/lib/backend.ts
      ↓
páginas e componentes React
```

O domínio não deve importar Wails, React, SQLite ou DTOs das APIs.

## Composição do backend

- `main.go`: inicialização Wails e assets incorporados.
- `app.go`: composition root, ciclo de vida e bindings públicos.
- `internal/domain`: entidades, contratos e tipos compartilhados pelos casos de uso.
- `internal/usecase`: regras do catálogo, builds, equipes, dano, rotações e IA.
- `internal/repository`: persistência SQLite.
- `internal/sources/nanoka`: cliente, DTOs e mapeamento da API Nanoka.
- A seleção de fonte e versão é persistida em `app_settings`. Nanoka Live 3.5
  e Latest 3.6.1 sincronizam snapshots fechados.
- `internal/sources/arikatsu`: adapter dos JSONs brutos das branches 3.5, 3.4 e
  3.3. Normaliza personagens, armas, Echoes, Sonata Effects e localização
  portuguesa para os mesmos DTOs internos usados pelo catálogo. Os arquivos
  brutos são armazenados em `sources/arikatsu/<versão>` no diretório local do
  app. Armas usam nativamente `weaponconf`, `weaponpropertygrowth`,
  `propertyindex` e `weaponreson`: ATK e atributo secundário são calculados no
  nível 90 e a história fica separada do efeito passivo. Campos detalhados ainda
  não presentes no adapter são enriquecidos pelo
  snapshot Nanoka da mesma versão; se o detalhe não estiver disponível, a
  sincronização preserva o detalhe anterior em vez de gravar dados vazios.
- Variantes cosméticas `MonsterInfo_35*` existem no inventário bruto Arikatsu,
  mas não são entradas independentes do Data Bank e não entram no catálogo de
  Echoes.
- A substituição do catálogo de Echoes remove registros fora do snapshot apenas
  quando não estão referenciados em `owned_echoes`; dados pessoais nunca são
  apagados pela troca de fonte ou versão.
- `internal/sources/guide`: cliente dos guias externos.
- `internal/sources/convene`: descoberta e decodificação do histórico local,
  leitura da URL de importação e catálogo dinâmico de banners.
- `internal/httpcache`: transporte HTTP persistente. Revalida GETs com
  `ETag`/`Last-Modified`, reaproveita respostas `304 Not Modified` e mantém a
  última resposta válida como fallback offline.
- `internal/assets`: cache validado de imagens.
- `internal/database`: abertura, backup, restauração e migrations incorporadas.
- `cmd/wavearchive-cli`: CLI que reutiliza o núcleo Go.

## Frontend

- `frontend/src/App.tsx`: shell, navegação e roteamento local.
- `frontend/src/lib/backend.ts`: única fronteira para os bindings Wails.
- `frontend/src/types.ts`: contratos consumidos pelo React.
- `frontend/src/CharacterDetail.tsx`: perfil, materiais, Forte, árvore e lore.
- `frontend/src/TeamsPage.tsx`: equipes, presets, biblioteca, tags oficiais e
  vínculo de uma Build compatível a cada posição.
- `frontend/src/BuildsPage.tsx`: composição de Builds, seleção de peças do
  catálogo ou do inventário e resumo reativo dos efeitos de Sonata.
- `frontend/src/LibraryFilterBar.tsx`: estrutura única de busca, ordenação e
  facetas usada pelas bibliotecas contextuais de Equipes e Builds.
- `frontend/src/AssistantPage.tsx`: chat contextual e configuração da sessão.
- `frontend/src/ConvenePage.tsx`: importação, filtros, pity, linha do tempo,
  inspeção de banners e exclusão do Histórico de Convene.
- `frontend/src/GlobalSearch.tsx`: pesquisa global local (`Ctrl+K`) entre
  personagens, habilidades, materiais, armas, Echoes, Sonatas, builds, equipes
  e conversas de IA. Habilidades e materiais são consultados no SQLite por
  `SearchCharacterContent` e abrem diretamente a aba correspondente do
  personagem.
- `frontend/src/WorkspacePages.tsx`: tela inicial operacional, configurações e
  conta; o Dashboard deriva recentes, favoritos e incompletos dos contratos
  existentes.
- `frontend/src/AdvancedFilters.tsx`: estrutura compartilhada dos filtros
  avançados e intervalos numéricos.
- `frontend/src/lib/navigation.ts`: entrega o item selecionado na pesquisa
  global à página de destino sem criar um segundo roteador.
- `frontend/src/lib/contextualShortcuts.ts`: aplica `Ctrl+N` e `Ctrl+S` somente
  nos compositores que registram ações contextuais e protege campos editáveis.
- `frontend/src/lib/sonata.ts`: deriva, sem inventar dados, a contagem e o estado
  dos efeitos a partir das peças escolhidas e das descrições sincronizadas.
- `frontend/src/lib/teamSynergy.ts`: consolida as tags oficiais dos membros da
  Equipe, mantendo descrição e personagens de origem.
- `frontend/src/styles.css`: tokens e estilos globais.

Páginas não devem chamar SQLite ou APIs externas diretamente.

Durante `CharacterCatalog.Sync`, os ícones de `profile.Skills` são baixados
depois dos detalhes e antes do `ReplaceSynced`. O mesmo caminho `/cache/` é
associado à habilidade e à entrada correspondente em
`profile.Progression.Skills`, evitando ícones genéricos quando a fonte fornece
o asset oficial.

O vínculo entre Equipe e Build usa exclusivamente `TeamMember.buildId`. A
interface filtra Builds pelo personagem do slot e o `TeamManager` repete essa
validação antes de persistir, impedindo referências cruzadas inválidas.

Os filtros avançados são aplicados no SQLite para catálogos oficiais e no
cliente apenas para coleções já carregadas, como Sonatas e equipes. Os campos
disponíveis devem refletir somente dados reais do domínio. Grades extensas usam
`content-visibility: auto` e tamanho intrínseco para evitar renderizar conteúdo
fora da área visível sem alterar a navegação ou o layout.

## Dados e arquivos locais

- Banco: `%AppData%\WaveArchive\wavearchive.db`.
- Assets: `%AppData%\WaveArchive\assets`.
- Cache HTTP: `%AppData%\WaveArchive\http-cache`.
- Backups: `%AppData%\WaveArchive\backups`.
- Snapshots: `%AppData%\WaveArchive\snapshots`.
- Build Windows: `build/bin/WaveArchive.exe`.

O Histórico de Convene é armazenado no SQLite local pelas tabelas da migration
`0016_convene_history.sql`. A importação é idempotente: registros já existentes
não são duplicados, e a exclusão remove somente o histórico importado.

SQLite usa foreign keys, WAL, busy timeout e migrations sequenciais em
`internal/database/migrations`. Dados oficiais e dados pessoais são mantidos
separados sempre que possível.

## Assistente IA

O assistente já possui:

- provedores Ollama, LM Studio e Gemini;
- streaming por evento Wails `ai:chunk`;
- conversas e mensagens persistidas no SQLite;
- contexto de personagem, build, equipe e rotação;
- busca RAG local por FTS5;
- guias oficiais sincronizados por personagem;
- modos estrito, assistido e geral.

Fluxo:

```text
AssistantPage
  → AssistantChatStream
  → AssistantService
  → contexto determinístico + histórico + fontes RAG
  → AIAnalyzer
  → provedor configurado
  → mensagem persistida + streaming para o React
```

O modo estrito não pode inventar informação ausente. Chaves remotas ficam
somente na sessão e nunca entram em logs.

## Regras de evolução

- Mudança de schema exige migration nova.
- Novo provedor de dados exige cliente/DTO/mapper próprios.
- Cálculos pertencem ao Go; React coleta escolhas e apresenta resultados.
- Bindings públicos devem ter tipos equivalentes em `frontend/src/types.ts` e
  funções em `frontend/src/lib/backend.ts`.
- Alterações estruturais devem atualizar este arquivo e `prompt/brain.md`.
