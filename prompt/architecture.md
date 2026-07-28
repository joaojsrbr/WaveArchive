# Arquitetura viva — WaveArchive

Última revisão: 2026-07-28.

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
  app. Campos detalhados ainda não presentes no adapter são enriquecidos pelo
  snapshot Nanoka da mesma versão; se o detalhe não estiver disponível, a
  sincronização preserva o detalhe anterior em vez de gravar dados vazios.
- Variantes cosméticas `MonsterInfo_35*` existem no inventário bruto Arikatsu,
  mas não são entradas independentes do Data Bank e não entram no catálogo de
  Echoes.
- A substituição do catálogo de Echoes remove registros fora do snapshot apenas
  quando não estão referenciados em `owned_echoes`; dados pessoais nunca são
  apagados pela troca de fonte ou versão.
- `internal/sources/guide`: cliente dos guias externos.
- `internal/assets`: cache validado de imagens.
- `internal/database`: abertura, backup, restauração e migrations incorporadas.
- `cmd/wavearchive-cli`: CLI que reutiliza o núcleo Go.

## Frontend

- `frontend/src/App.tsx`: shell, navegação e roteamento local.
- `frontend/src/lib/backend.ts`: única fronteira para os bindings Wails.
- `frontend/src/types.ts`: contratos consumidos pelo React.
- `frontend/src/CharacterDetail.tsx`: perfil, materiais, Forte, árvore e lore.
- `frontend/src/TeamsPage.tsx`: equipes, presets, biblioteca e tags oficiais.
- `frontend/src/AssistantPage.tsx`: chat contextual e configuração da sessão.
- `frontend/src/styles.css`: tokens e estilos globais.

Páginas não devem chamar SQLite ou APIs externas diretamente.

## Dados e arquivos locais

- Banco: `%AppData%\WaveArchive\wavearchive.db`.
- Assets: `%AppData%\WaveArchive\assets`.
- Backups: `%AppData%\WaveArchive\backups`.
- Snapshots: `%AppData%\WaveArchive\snapshots`.
- Build Windows: `build/bin/WaveArchive.exe`.

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
