# WaveArchive

WaveArchive é um aplicativo desktop local-first para consultar dados de
Wuthering Waves e organizar personagens, armas, Echoes, builds e equipes.

O projeto usa **Go 1.23**, **Wails v2**, **React 19**, **TypeScript**, **Vite** e
**SQLite**. Os dados sincronizados e as informações pessoais permanecem no
computador do usuário.

## Funcionalidades

- catálogo de personagens, armas, Echoes e Sonata Effects;
- pesquisa global com `Ctrl+K`, filtros avançados, favoritos e ordenação
  original da API;
- perfil completo do personagem com kit, árvore, sequências, materiais,
  atributos e lore;
- compositor de builds com personagem, arma e até cinco Echoes;
- limite de custo 12 e configuração de atributo principal e subatributos;
- equipes de três personagens e presets baseados nos dados sincronizados;
- conta local, importação, exportação e backups;
- calculadora determinística e dados de inimigos;
- assistente contextual com Ollama, LM Studio e Gemini;
- sincronização com cache HTTP condicional (ETag/Last-Modified), fallback
  offline e cache local de dados e imagens;
- seleção de fonte entre Nanoka (Live 3.5 ou Latest 3.6.1) e Arikatsu Data
  (branches 3.5, 3.4 e 3.3).

Dados oficiais exibidos pelo aplicativo vêm das fontes sincronizadas. Dados
ausentes não devem ser inventados.

## Estrutura

```text
wave_archive/
├── cmd/                  CLI auxiliar
├── docs/                 documentação técnica
├── frontend/             aplicação React
├── internal/             domínio, casos de uso, banco e integrações
├── prompt/               contexto essencial para outras IAs
├── app.go                composição e bindings Wails
├── main.go               inicialização desktop
└── wails.json            configuração do Wails
```

O frontend acessa o backend somente pelos bindings definidos em
`frontend/src/lib/backend.ts`. Regras de domínio, persistência e cálculos ficam
no Go.

## Desenvolvimento

Requisitos:

- Go 1.23 ou superior;
- Node.js e npm;
- Wails CLI v2;
- WebView2 no Windows.

Instale as dependências do frontend:

```powershell
cd D:\projetos\wave_archive\frontend
npm install
```

Execute o aplicativo em desenvolvimento:

```powershell
cd D:\projetos\wave_archive
wails dev
```

## Validação

```powershell
go test ./...
npm --prefix frontend test
npm --prefix frontend run build
```

## Gerar o executável

```powershell
cd D:\projetos\wave_archive
wails build
```

O executável é gerado em `build/bin/WaveArchive.exe`.

## Dados locais

Por padrão, o WaveArchive mantém seus arquivos em:

```text
%AppData%\WaveArchive\
├── wavearchive.db
├── assets\
├── backups\
├── http-cache\
└── snapshots\
```

O arquivo `wuwa_scraper.py` é mantido somente como referência funcional e não é
executado pelo aplicativo.

## CLI auxiliar

```powershell
go run ./cmd/wavearchive-cli sync
go run ./cmd/wavearchive-cli --source arikatsu --version 3.5 sync
go run ./cmd/wavearchive-cli list
go run ./cmd/wavearchive-cli list-weapons
go run ./cmd/wavearchive-cli show 1102
```

## Continuidade com IA

Entregue primeiro [prompt/start.md](prompt/start.md) para a IA que trabalhará
no projeto. Ela direciona a leitura de:

- [prompt/architecture.md](prompt/architecture.md): arquitetura e contratos;
- [prompt/brain.md](prompt/brain.md): estado atual, decisões e riscos.

Esses são os únicos arquivos obrigatórios em `prompt/`.

## Documentação

- [Arquitetura inicial](docs/architecture.md)
- [Plano de implementação](docs/implementation-plan.md)
- [Design system](docs/design-system-v2.md)
- [Contrato entre interface e dados](docs/ui-data-contract.md)
