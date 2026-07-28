# Brain — memória operacional do WaveArchive

Última atualização: 2026-07-28.

Este arquivo registra decisões duráveis e o estado útil para a próxima IA.
Não substitui o código, não contém segredos e não deve virar um diário extenso.

## Objetivo atual

Manter um aplicativo desktop local-first, confiável e visualmente premium para
consultar dados oficiais, planejar progressão, montar builds e equipes, calcular
resultados e usar IA contextual sem inventar dados do jogo.

## Decisões confirmadas

- Usar Go + Wails + React + TypeScript + SQLite.
- O frontend não implementa regras de cálculo que pertencem ao domínio.
- Dados visíveis devem vir da API, de cálculos determinísticos ou do usuário.
- O assistente suporta Ollama, LM Studio e Gemini (Gemini nunca deve ser descrito como local).
- Chaves de API não são persistidas pelo chat.
- Banco, cache, backups e snapshots ficam no diretório de dados do WaveArchive.
- O design atual usa navegação superior, superfícies escuras minerais, ciano
  técnico e dourado para seleção (direção nanoka.cc).

## Estado implementado

- Catálogos completos (personagens, armas, Echoes, Sonata) via Arikatsu Data/Nanoka.
- Perfil detalhado: kit, sequências, materiais, Forte, lore, tags e árvore.
- Builds e Equipes com compositor contextual (máx 12 de custo nos Echoes, validações estritas).
- Dashboard, conta, importação, exportação e backups.
- Assistente contextual com histórico, streaming e RAG FTS5.
- Build Windows gerada em `build/bin/WaveArchive.exe`.
- Formatação automática rigorosa configurada nativamente via Wails hooks, VS Code e pre-commit Git.

## Arquivos sensíveis à arquitetura

- `app.go`: concentra a composição e os bindings.
- `internal/database/migrations`: nunca reescrever migrations aplicadas.
- `internal/domain`: deve permanecer independente de infraestrutura.
- `frontend/src/types.ts` & `frontend/src/lib/backend.ts`: fronteira e tipos do Wails.
- `wails.json` & `scripts/format.go`: orquestram processos críticos de build/format.

## Riscos e bloqueios conhecidos

- A documentação antiga em `docs/architecture.md` pode estar atrás do estado atual.
- O repositório local aparece com muitos arquivos ainda não rastreados, não apagar assumindo que são temporários.
- Mudanças paralelas no build e em `go test` podem disputar `frontend/dist`.
- O Wails (`wails build`) executa hooks de pré-build a partir da subpasta `build/bin/`, exigindo que scripts (`format.go`) localizem dinamicamente a raiz do projeto (via `go.mod`).

## Próximos trabalhos conhecidos

- Embeddings e ferramentas de IA com confirmação antes de gravar.
- Inventário de materiais e desconto do que o usuário já possui.
- ETag/Last-Modified para sincronização HTTP.
- Virtualização de listas grandes.

## Registro recente (Resumo Condensado)

### 2026-07-28 — Refatorações Maiores de UI e Catálogos
- Remoção do Planejador antigo, Convene Tracker, Rotações e Comparador (mantidos apenas no banco/arquivos para compatibilidade).
- Compositor de Builds e Equipes unificados (arquivo lateral, seleção central, inspetor contextual).
- Redesign do Kit & Árvore inspirado na nanoka.cc: progressão vertical por habilidade e cards, mapa radial com navegação lateral (Resonance Lens).
- Integração refinada do Arikatsu Data (versões legadas suportadas, variantes Rover agrupadas).
- Padronização de gutters, scrolls e paddings em todas as páginas para máxima consistência visual.

### 2026-07-28 — Ferramentas de Código e Auto-Format
- Adicionado Prettier (`frontend`) e configurado formatação no VS Code (`settings.json` e `extensions.json`).
- Criado `scripts/format.go` robusto, acionado via `preBuildHooks` no `wails.json` (formatando frontend e backend automaticamente no `wails build`, não importando o diretório de execução).
- Adicionado hook de Git `pre-commit` para garantir formatação limpa e transparente antes de cada commit.

## Como atualizar este arquivo

Ao terminar uma tarefa, altere apenas o necessário:
1. Atualize "Estado implementado" se uma capacidade mudou.
2. Mantenha os registros recentes resumidos para evitar gigantismo, agrupando pequenas atualizações e condensando históricos massivos.
3. Adicione ou remova riscos/próximos trabalhos conforme necessário.
