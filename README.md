# WaveArchive

Aplicativo desktop local-first para consultar dados de **Wuthering Waves**,
organizar a conta e montar personagens, Builds e Equipes sem depender de um
serviço remoto para armazenar informações pessoais.

**Versão atual: 1.0.0**

[Releases](https://github.com/joaojsrbr/WaveArchive/releases) ·
[Documentação](docs/) ·
[Plano do produto](docs/implementation-plan.md)

## Principais recursos

### Arquivo sincronizado

- catálogos de personagens, armas, Echoes e Sonata Effects;
- fontes selecionáveis Nanoka e Arikatsu Data;
- cache HTTP com `ETag`, `Last-Modified` e fallback offline;
- pesquisa global com `Ctrl+K` entre catálogos, habilidades, materiais e dados
  pessoais, além de filtros contextuais, favoritos e ordenação da fonte;
- agrupamento das variantes masculina e feminina do Rover.

### Personagens e progressão

- visão geral, Kit & Árvore, sequências, materiais, atributos e lore;
- ícones oficiais das habilidades e dados provenientes da fonte sincronizada;
- cálculo de materiais de ascensão e habilidades;
- registro local de posse, nível e sequência.

### Builds e inventário

- personagem, arma e até cinco Echoes, respeitando o custo máximo de 12;
- peças do inventário com nível, Sonata, atributo principal e subatributos;
- indicação de conjuntos de Sonata ativos e incompletos com descrição oficial;
- histórico de versões, favoritos, bloqueio, duplicação e restauração;
- atalhos contextuais `Ctrl+N` para criar e `Ctrl+S` para salvar.

### Equipes

- composição com três personagens;
- vínculo de uma Build compatível a cada posição;
- presets provenientes dos dados sincronizados;
- resumo de sinergia baseado somente nas tags oficiais da fonte.

### Histórico de Convene

- importação automática pelo log local do jogo, seleção manual do log ou URL;
- banners carregados dinamicamente a partir dos dados oficiais;
- histórico pesquisável com raridade, tipo, data e pity;
- linha do tempo dos resultados de cinco estrelas;
- associação com personagens e armas do catálogo para reutilizar os ícones;
- prevenção de registros duplicados e exclusão completa do histórico importado.

### Conta, segurança e IA

- banco SQLite local, backups, snapshots, importação e exportação do workspace;
- tela inicial com itens recentes, favoritos e composições incompletas;
- assistente contextual com Ollama, LM Studio e Gemini;
- nenhum dado ausente é apresentado como oficial ou inventado pelo frontend.

## Instalação no Windows

1. Abra a página de
   [Releases](https://github.com/joaojsrbr/WaveArchive/releases).
2. Escolha a versão desejada.
3. Baixe `WaveArchive.exe` quando o executável estiver anexado à release.
4. Execute o aplicativo.

O Windows precisa do **WebView2 Runtime**, normalmente já instalado em versões
atuais do sistema.

## Dados locais

O WaveArchive mantém dados pessoais, catálogo sincronizado e caches no
computador do usuário:

```text
%AppData%\WaveArchive\
├── wavearchive.db
├── assets\
├── backups\
├── http-cache\
├── snapshots\
└── sources\
```

O Histórico de Convene importado é persistido no mesmo banco local. Use a ação
de exclusão da própria tela para removê-lo.

## Stack

- Go 1.23;
- Wails v2;
- React 19;
- TypeScript e Vite;
- SQLite via `modernc.org/sqlite`.

O frontend acessa o backend exclusivamente pelos bindings centralizados em
`frontend/src/lib/backend.ts`. Persistência, normalização e regras
determinísticas ficam no Go.

## Desenvolvimento

Requisitos:

- Go 1.23 ou superior;
- Node.js e npm;
- Wails CLI v2;
- WebView2 no Windows.

Instale as dependências:

```powershell
cd frontend
npm install
cd ..
```

Execute em desenvolvimento:

```powershell
wails dev
```

Valide o projeto:

```powershell
go test ./...
npm --prefix frontend test
npm --prefix frontend run build
```

Gere o executável:

```powershell
wails build
```

O resultado é criado em `build/bin/WaveArchive.exe`.

## CLI auxiliar

```powershell
go run ./cmd/wavearchive-cli sync
go run ./cmd/wavearchive-cli --source arikatsu --version 3.5 sync
go run ./cmd/wavearchive-cli list
go run ./cmd/wavearchive-cli list-weapons
go run ./cmd/wavearchive-cli show 1102
```

## Estrutura do repositório

```text
wave_archive/
├── cmd/                  CLI auxiliar
├── docs/                 documentação técnica e de produto
├── frontend/             interface React
├── internal/             domínio, casos de uso, fontes e persistência
├── prompt/               contexto essencial para continuidade com IA
├── app.go                composição e bindings Wails
├── main.go               inicialização desktop
└── wails.json            configuração do aplicativo
```

Para continuar o trabalho com outra IA, entregue primeiro
[`prompt/start.md`](prompt/start.md). Ele direciona a leitura da arquitetura e
do estado atual sem exigir todos os documentos do repositório.

## Documentação

- [Arquitetura viva](prompt/architecture.md)
- [Memória operacional](prompt/brain.md)
- [Plano de implementação](docs/implementation-plan.md)
- [Design system](docs/design-system-v2.md)
- [Contrato entre interface e dados](docs/ui-data-contract.md)
- [Changelog](CHANGELOG.md)
