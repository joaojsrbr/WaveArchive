# Plano de implementação — WaveArchive

Última revisão: 2026-07-29.

Este plano registra apenas o produto atual. Features removidas da interface não
devem voltar por acidente só porque tabelas ou estruturas antigas ainda existem.

## Fase 0 — fundação concluída

- [x] Leitura de `prompt/start.md`, `prompt/architecture.md` e
  `prompt/brain.md`.
- [x] Stack Go, Wails, React, TypeScript e SQLite.
- [x] Arquitetura em domínio, casos de uso, repositórios e fontes externas.
- [x] SQLite com WAL, migrations incorporadas, FTS5, backup e restauração.
- [x] Design system escuro e responsivo inspirado em Wuthering Waves.

## Fase 1 — shell, navegação e produtividade

- [x] Navegação superior responsiva e breadcrumbs.
- [x] Pesquisa global com `Ctrl+K`.
- [x] Persistência local de filtros, visualização e última página.
- [x] Estados de carregamento, vazio, erro e sincronização.
- [x] Ação contextual de `Ctrl+N`:
  - Equipes cria uma nova equipe;
  - Builds cria uma nova build;
  - demais páginas não executam uma criação implícita.
- [x] Atalho contextual de `Ctrl+S` para salvar a composição aberta.
- [ ] Logs em arquivo com rotação, exportação e tela de diagnóstico.

## Fase 2 — fontes e sincronização

- [x] Nanoka como catálogo normalizado.
- [x] Arikatsu Data como fonte selecionável, incluindo personagens, armas e
  Echoes normalizados.
- [x] Seleção de versão e suporte à versão mais recente e às duas anteriores
  quando disponíveis na fonte.
- [x] Retry transitório, timeout, cancelamento e limite de resposta.
- [x] Upsert transacional sem sobrescrever posse, favoritos ou outros dados
  pessoais.
- [x] Snapshot antes da sincronização e restauração segura.
- [x] Cache HTTP persistente com `ETag`, `Last-Modified`, respostas `304` e
  fallback offline.
- [x] Download e cache atômico de imagens.
- [x] Progresso detalhado da sincronização na interface.
- [ ] Tela de diferenças da atualização antes de aplicar uma nova versão:
  personagens, armas, Echoes e Sonatas adicionados, alterados ou removidos.
- [ ] Política configurável de retenção para cache, imagens e snapshots.

## Fase 3 — catálogos

- [x] Personagens, armas, Echoes e Sonata Effects.
- [x] Busca, ordenação da API, filtros por dados reais e grade/tabela quando
  aplicável.
- [x] Rover masculino e feminino agrupados por atributo.
- [x] Posse, nível, sequência, refinamento e favoritos.
- [x] Página de personagem com visão geral, Kit & Árvore, sequências,
  materiais, atributos e lore.
- [x] Ícones oficiais das habilidades baixados durante a sincronização e
  reutilizados entre o Kit, detalhes e progressão.
- [x] Arma assinatura quando existir e arma recomendada para Rover.
- [x] Inventário local de Echoes com atributo principal e subatributos válidos.
- [x] Renderização diferida de grades extensas com `content-visibility`.
- [ ] Paginação ou virtualização real somente se medições mostrarem que a
  renderização diferida deixou de ser suficiente.
- [ ] Inventário de materiais do usuário integrado aos totais de ascensão.

## Fase 4 — Builds

- [x] CRUD, duplicação, favorito, bloqueio, exclusão e restauração.
- [x] Personagem e arma compatível na composição.
- [x] Até cinco Echoes respeitando custo total máximo de 12.
- [x] Nível do Echo, Sonata, atributo principal e cinco subatributos.
- [x] Arma assinatura priorizada; Rover usa recomendação em vez de assinatura.
- [x] Biblioteca vertical com busca, ordenação e filtros contextuais:
  - personagem não mostra filtro de arma;
  - arma não mostra filtros de personagem;
  - Echo mostra custo, chips de Sonata e seus próprios atributos.
- [x] Histórico automático de versões da build.
- [x] Importação e exportação pelo arquivo portátil do workspace.
- [x] Validação visual da composição de Sonata, mostrando conjuntos ativos e
  incompletos com a descrição oficial de cada efeito.
- [x] Reutilizar diretamente peças cadastradas no inventário de Echoes, sem
  copiar manualmente seus subatributos.

## Fase 5 — Equipes

- [x] CRUD, duplicação, exclusão e restauração.
- [x] Composição com três personagens.
- [x] Presets provenientes de guias sincronizados.
- [x] Tags e funções oficiais da API, sem funções inventadas.
- [x] Biblioteca compartilhada com Builds e filtros coerentes.
- [x] Uma única rolagem na página; arquivo, inspetor e biblioteca não possuem
  scroll vertical independente.
- [x] Ação `Nova equipe` no topo do arquivo lateral.
- [x] Vincular uma build salva a cada posição da equipe, limitada ao mesmo
  personagem e com estado explícito quando a Build fica indisponível.
- [x] Resumo de sinergia baseado exclusivamente nas tags oficiais dos três
  personagens, com indicação navegável da fonte de cada informação.

## Fase 6 — IA

- [x] Ollama e LM Studio locais.
- [x] Gemini remoto com chave somente em memória.
- [x] Modos estrito, assistido e geral.
- [x] Chat contextual para personagem, build e equipe.
- [x] Streaming, teste de conexão, descoberta de modelos e RAG com FTS5.
- [ ] Embeddings locais opcionais.
- [ ] Ferramentas de IA capazes de propor alterações em builds e equipes, mas
  sempre exigindo confirmação antes de gravar.

## Fase 7 — confiabilidade e entrega

- [x] Backup, restauração e arquivo portátil validado.
- [x] Build Windows em `build/bin/WaveArchive.exe`.
- [x] Formatação automática no Wails, VS Code e pre-commit.
- [x] Testes dos repositórios e regras determinísticas principais.
- [ ] Automação de fluxos do aplicativo Wails empacotado.
- [ ] Testes de acessibilidade para teclado, foco, zoom e redução de movimento.
- [ ] Relatório de integridade da fonte mostrando campos ausentes, assets
  quebrados e registros descartados durante a normalização.

## Fora do escopo atual

As seguintes features foram removidas da interface por decisão de produto:

- Planejador antigo.
- Convene Tracker.
- Rotações.
- Comparador.

Estruturas legadas podem continuar no banco apenas para compatibilidade de
migrations e arquivos antigos, mas não autorizam o retorno dessas telas.

## Fase 8 — features aprovadas

### 8.1 — atalhos contextuais

- [x] `Ctrl+N` cria uma nova equipe em Equipes e uma nova build em Builds.
- [x] `Ctrl+S` salva a equipe ou build aberta quando a composição for válida.
- [x] Não executar as ações do app enquanto o foco estiver em campos de texto,
  selects ou áreas editáveis.
- [x] Exibir confirmação visual curta após criar, salvar ou impedir uma ação
  inválida.

### 8.2 — inventário de Echoes integrado às Builds

- [x] Adicionar uma origem `Inventário` na biblioteca de Echoes da Build.
- [x] Permitir escolher uma peça real já cadastrada, preservando ID, nível,
  Sonata, atributo principal e subatributos.
- [x] Indicar quando uma peça do inventário já está sendo usada em outra build,
  sem impedir reutilização.
- [ ] Manter o snapshot dos atributos dentro do histórico da build para que
  versões antigas não mudem quando a peça for editada depois.
- [ ] Se a peça for removida do inventário, preservar a Build e marcar sua
  origem como indisponível.
- [x] Reaproveitar as tabelas e contratos de `OwnedEcho`; criar migration apenas
  se os testes mostrarem que a referência atual não é persistida com segurança.

### 8.3 — conjuntos de Sonata na Build

- [x] Calcular a contagem por Sonata usando somente as peças selecionadas.
- [x] Mostrar conjuntos ativos e o número de peças restante para cada efeito
  incompleto.
- [x] Destacar quando as peças de um conjunto ainda não ativam nenhum efeito.
- [x] Ler da fonte sincronizada as descrições oficiais dos efeitos de duas e
  cinco peças, sem gerar texto ou bônus no frontend.
- [x] Atualizar o resumo imediatamente ao adicionar, remover ou trocar a Sonata
  de um Echo.
- [x] Manter cor, texto e contagem para que o estado não dependa somente
  de cor.

### 8.4 — Builds vinculadas aos personagens da Equipe

- [x] Usar o `buildId` já existente em `TeamMember`, sem criar relação paralela.
- [x] Na posição selecionada, listar somente Builds do mesmo personagem.
- [x] Permitir vincular, substituir, abrir e desvincular a Build.
- [x] Exibir um resumo compacto de arma, custo dos Echoes e Sonatas.
- [x] Se a Build for excluída, manter o personagem e mostrar `Build
  indisponível` até o usuário escolher outra.

### 8.5 — sinergia oficial da Equipe

- [x] Consolidar as tags oficiais dos três personagens selecionados.
- [x] Separar funções compartilhadas e individuais sem criar pontuação ou
  inferir capacidades ausentes que a fonte não declarou.
- [x] Exibir a origem de cada tag e permitir selecionar o personagem
  correspondente.
- [x] Usar os textos sincronizados da API/guia; não gerar recomendações como se
  fossem dados oficiais.
- [x] Atualizar o resumo imediatamente ao trocar um personagem.

### 8.6 — tela inicial operacional

- [x] Seção `Recentes` com Builds e Equipes ordenadas pela última alteração.
- [x] Seção `Favoritos` reunindo itens favoritos dos catálogos e composições.
- [x] Seção `Incompletos` com Builds sem arma ou com menos de cinco Echoes e
  Equipes sem três Builds válidas.
- [x] Cada item abre diretamente a tela e o registro correto.
- [x] Estado vazio útil, sem cards decorativos ou métricas inventadas.
- [x] Limitar a quantidade inicial para manter a tela inicial rápida.

## Ordem de implementação aprovada

1. Atalhos contextuais `Ctrl+N` e `Ctrl+S`.
2. Inventário de Echoes integrado às Builds.
3. Indicadores de Sonata ativos e incompletos.
4. Builds vinculadas às posições da Equipe.
5. Sinergia baseada nas tags oficiais.
6. Tela inicial de recentes, favoritos e incompletos.

A tela inicial fica por último porque depende dos estados produzidos pelas
outras features. O vínculo de Build usa o campo `buildId` já existente, e a
sinergia deve consumir somente dados oficiais que já passaram pela
normalização.

## Features não selecionadas nesta rodada

- Inventário de materiais com cálculo do que ainda falta.
- Comparação entre versões antes da sincronização.
- Relatório de integridade das fontes Nanoka e Arikatsu.
- Compartilhamento individual de Build/Equipe por arquivo ou código.

## Bloqueios e regras permanentes

- Não inventar dados, funções, recomendações ou relações que não existam nas
  fontes sincronizadas.
- Preservar dados pessoais durante troca de fonte ou versão.
- Não reintroduzir features listadas como fora do escopo sem nova decisão
  explícita.
- Antes de virtualização complexa, medir tempo de renderização, memória e
  fluidez com o catálogo completo.
- Manter testes para bancos legados sem `schema_migrations`.
