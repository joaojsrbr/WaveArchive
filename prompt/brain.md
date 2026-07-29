# Brain — memória operacional do WaveArchive

Última atualização: 2026-07-29.

Este arquivo registra decisões duráveis e o estado útil para a próxima IA.
Não substitui o código, não contém segredos e não deve virar um diário extenso.

## Objetivo atual

Manter um aplicativo desktop local-first, confiável e visualmente premium para consultar dados oficiais, planejar progressão, montar builds/equipes e usar IA contextual sem inventar dados do jogo.

## Decisões confirmadas

- Stack: Go + Wails + React + TypeScript + SQLite.
- Frontend não implementa regras de domínio; dados vêm da API/cálculos/usuário.
- Assistente suporta Ollama, LM Studio e Gemini (não descrito como local). Sem persistir chaves.
- Banco, cache, backups no diretório de dados local.
- UI: navegação superior, superfícies escuras minerais, ciano técnico e dourado (direção nanoka.cc).

## Estado implementado

- Catálogos via Arikatsu Data/Nanoka (personagens, armas, Echoes, Sonata).
- Perfil detalhado (kit, sequências, materiais, Forte, lore, tags, árvore).
- Ícones oficiais das skills são convertidos dos caminhos Nanoka `/Game/Aki/...`
  para assets locais em `/cache/characters/<id>/skills/`.
- Builds e Equipes: compositor contextual unificado, validações estritas (máx
  12 de custo), atalhos contextuais `Ctrl+N`/`Ctrl+S` e abas sem scroll duplo.
- Builds podem selecionar uma peça real do inventário de Echoes, preservando
  sua identidade e atributos. O resumo de Sonata mostra efeitos ativos e
  incompletos usando as descrições oficiais da fonte sincronizada.
- Cada posição de uma Equipe pode referenciar uma Build salva do mesmo
  personagem. O vínculo usa `TeamMember.buildId`, mostra arma, custo, Echoes e
  Sonatas, e preserva o personagem quando a Build fica indisponível.
- Equipes consolidam as tags oficiais dos três personagens em funções
  compartilhadas e individuais, preservando descrição e origem sem gerar
  pontuação. A tela inicial reúne recentes, favoritos e composições incompletas
  usando somente o estado persistido.
- Funcionalidades gerais: Dashboard, conta, import/export, backups, busca global (`Ctrl+K`), filtros avançados consistentes em todas as entidades.
- Assistente contextual: histórico, streaming, RAG FTS5.
- Cache HTTP condicional com ETag/Last-Modified. Renderização diferida em grades extensas.
- Formatação automática nativa via Wails hooks, VS Code e pre-commit Git.
- Build Windows: `build/bin/WaveArchive.exe`.

## Arquivos sensíveis à arquitetura

- `app.go`: composição e bindings.
- `internal/database/migrations`: nunca edite migrations aplicadas.
- `internal/domain`: independente de infraestrutura.
- `frontend/src/types.ts` & `frontend/src/lib/backend.ts`: contratos Wails.
- `wails.json` & `scripts/format.go`: build e hooks de formatação.

## Riscos e bloqueios conhecidos

- `docs/architecture.md` pode estar desatualizado.
- O Wails (`wails build`) roda hooks a partir de `build/bin/`, exigindo que scripts localizem dinamicamente a raiz. Arquivos não rastreados podem não ser apenas temporários.

## Próximos trabalhos conhecidos

- Embeddings e ferramentas de IA (com confirmação).
- Inventário de materiais (desconto do existente).
- Avaliar paginação se a renderização diferida não for suficiente no futuro.

## Registro recente condensado

- **Jul 2026:**
  - Busca global (`Ctrl+K`), filtros avançados coesos, chips contextuais e sincronização eficiente HTTP.
  - Otimização de renderização diferida e scrolls consistentes em Builds/Equipes.
  - Refatoração de UI (novo compositor unificado, redesign do Kit/Árvore).
  - Integração do inventário real de Echoes às Builds, resumo oficial de Sonata
    e atalhos contextuais de criação/salvamento.
  - Vínculo de Builds salvas às posições das Equipes, com validação por
    personagem no backend.
  - Sinergia oficial rastreável e tela inicial operacional baseada no workspace.
  - Cache local dos ícones oficiais de Normal Attack, Resonance Skill,
    Liberation, Forte, Intro, Outro e habilidades inerentes.
  - Integração de auto-formatação no Wails e hooks do Git.

## Como atualizar este arquivo

Altere apenas o necessário, mantenha "Registro recente" agrupado e muito condensado, evite adicionar logs extensos.
