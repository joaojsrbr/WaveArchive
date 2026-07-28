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
- A ordenação da API é preservada em `characters.api_order`.
- Tags de personagens vêm de `CharacterExtras`, sincronizadas da API.
- Equipes possuem três personagens distintos.
- Presets exibem sua fonte; dados de usuário e dados sincronizados são distintos.
- O assistente suporta Ollama, LM Studio e Gemini.
- Gemini é remoto e nunca deve ser descrito como IA local.
- Chaves de API não são persistidas pelo chat.
- Banco, cache, backups e snapshots ficam no diretório de dados do WaveArchive.
- O design atual usa navegação superior, superfícies escuras minerais, ciano
  técnico e dourado para seleção.

## Estado implementado

- Catálogos de personagens, armas, Echoes e Sonata.
- Perfil detalhado com kit, sequências, materiais, Forte, lore, tags e árvore.
- Kit & Árvore em formato editorial inspirado na nanoka.cc: progressão por
  habilidade, bônus vinculados e cards detalhados alimentados pelo perfil sincronizado.
- Planejador de progressão com custos oficiais.
- Builds, equipes, buffs, dano e DPS.
- Lista de equipes e presets com compositor de três slots.
- Biblioteca de personagens em grade vertical.
- Inspetor de tags com rolagem própria.
- Dashboard, conta, importação, exportação e backups.
- Assistente contextual com histórico, streaming e RAG FTS5.
- Build Windows gerada em `build/bin/WaveArchive.exe`.

## Arquivos sensíveis à arquitetura

- `app.go`: concentra a composição e os bindings; evitar transformá-lo em
  repositório ou caso de uso.
- `internal/database/migrations`: nunca reescrever migrations aplicadas.
- `internal/domain`: deve permanecer independente de infraestrutura.
- `frontend/src/types.ts`: manter sincronizado com os tipos expostos pelo Go.
- `frontend/src/lib/backend.ts`: manter como fronteira do Wails no React.
- `wuwa_scraper.py`: referência somente leitura, não runtime.

## Riscos e bloqueios conhecidos

- A documentação antiga em `docs/architecture.md` descreve a primeira fatia
  vertical e pode estar atrás do estado atual; este brain registra o panorama
  mais recente.
- O repositório local aparece com muitos arquivos ainda não rastreados. Não
  apagar, mover ou sobrescrever esses arquivos por assumir que são temporários.
- Mudanças paralelas no build e em `go test` podem disputar `frontend/dist`;
  execute essas etapas sequencialmente.
- O cache Go padrão do usuário pode negar acesso no ambiente Codex. Quando
  necessário, use um `GOCACHE` temporário dentro do projeto e remova-o ao final.

## Próximos trabalhos conhecidos

- Embeddings e ferramentas de IA com confirmação antes de gravar.
- Inventário de materiais e desconto do que o usuário já possui.
- ETag/Last-Modified para sincronização HTTP.
- Virtualização de listas grandes.
- Automação do aplicativo Wails empacotado.

## Registro recente

### 2026-07-28 — remoção de Planejador e Convene Tracker

- Planejador de metas e Convene Tracker foram removidos da navegação e do frontend.
- Os bindings públicos dessas duas features também foram removidos do Wails.
- Dashboard deixou de exibir metas e pity.
- Tabelas e dados antigos permanecem no banco e no formato de arquivo apenas para
  compatibilidade; nenhuma migration aplicada foi reescrita e nenhum dado foi apagado.

### 2026-07-28 — remoção de Rotações e Comparador
- Rotações e Comparador foram removidos da navegação e do frontend.
- Os bindings públicos para criar, excluir e avaliar rotações foram removidos do Wails.
- Referências às duas features também saíram de Builds, Equipes e Assistente IA.
- Estruturas e dados antigos de rotações permanecem no banco e no arquivo portátil
  apenas para compatibilidade; nenhuma migration foi reescrita e nenhum dado foi apagado.

### 2026-07-28 — novo compositor de Builds
- A página de Builds segue a linguagem do compositor de Equipes: arquivo lateral,
  composição central, inspetor contextual e biblioteca horizontal.
- O fluxo permite selecionar personagem, arma compatível e até cinco Echoes.
- O custo total dos Echoes não pode ultrapassar 12, com validação no React e no Go.
- Cada Echo pode configurar nível, Sonata, atributo principal e cinco subatributos
  antes de ser registrado no inventário e vinculado à Build.
- A biblioteca usa grade vertical com rolagem própria e filtros contextuais avançados.
- Personagens filtram por elemento, raridade, arma, posse e favoritos.
- Armas filtram por raridade, posse e favoritos; a assinatura oficial do personagem
  fica sempre em primeiro quando `signatureWeapon` estiver presente no perfil.
- Echoes filtram por custo, Sonata e classe.

### 2026-07-28 — biblioteca e editor de Echo refinados

- A página de Builds usa um único scroll vertical no documento. Biblioteca, inspetor
  e arquivo lateral fluem com a página, sem regiões internas roláveis.
- O filtro de Sonata passou de `select` para chips com nomes e ícones reais
  sincronizados da API; os chips quebram em múltiplas linhas, sem scroll horizontal.
- A ação `Nova build` fica no topo do arquivo lateral, antes da busca e das builds.
- O inspetor de Echo segue a hierarquia visual de uma ficha de equipamento: retrato,
  nível, Sonata, atributo principal e rolagens em linhas compactas.
- Subatributos deixaram de ser texto livre. O usuário escolhe uma das 13 categorias
  reais e uma rolagem discreta válida; categorias duplicadas no mesmo Echo ficam
  indisponíveis.
- Um slot de subatributo é liberado a cada +5 níveis, até cinco no +25. Ao reduzir
  o nível, rolagens de slots bloqueados são removidas.
- Os valores foram transcritos das tabelas de Echo Stats do Wuthering Waves Wiki e
  conferidos contra os intervalos documentados pela Prydwen. A persistência continua
  usando strings legíveis em `substatsJson`, preservando compatibilidade com o avaliador.
- O compositor e o inspetor não possuem padding externo próprio; o espaçamento fica
  nos blocos internos, mantendo as bordas principais simétricas em diferentes larguras.
- Inspetores de personagem e arma usam 14 px de padding interno uniforme; o painel
  externo continua sem padding para preservar a simetria da malha.
- Os selects de subatributo e valor forçam esquema escuro também nas opções nativas.
- Variantes `Rover:*` não exibem arma “Assinatura”. A primeira arma recomendada pela
  API permanece priorizada, mas recebe o rótulo correto “Recomendada” na Build e no
  detalhe do personagem.

### 2026-07-28 — Kit & Árvore / direção nanoka.cc

- O mapa radial foi descartado por decisão do usuário.
- A referência visual passou a ser a página de personagem da nanoka.cc:
  superfícies roxo-escuras, progressão por habilidade e cards editoriais.
- Bônus de atributo aparecem dentro da habilidade-mãe, nunca como habilidades
  independentes; a relação usa `parentNodes` e `nodeType`.
- Modos, níveis, valores, descrições e entradas do Forte continuam usando
  exclusivamente o perfil sincronizado.
- O fundo raster da direção radial rejeitada foi removido do projeto e do bundle.
- A antiga paleta roxa editorial foi removida por decisão do usuário. Kit & Árvore
  agora reutiliza a paleta da Visão Geral: fundo `#090e13`, painéis `#0d141a`,
  bordas azul-acinzentadas, ciano `#67dce3` e dourado `#e8bb58`.

### 2026-07-28 — Kit & Árvore / Resonance Lens restaurado

- A composição vertical Forte Meridian foi revertida conforme decisão do usuário.
- Restaurado o navegador lateral, mapa radial e inspetor contextual da versão anterior.
- O hero do personagem permanece visível em formato compacto para priorizar o mapa.
- Refinados atmosfera, contraste, seleção, modos e densidade sem criar dados;
  habilidades, relações, níveis e modos continuam vindo do perfil sincronizado.
- Busca, troca de modo e seleção de nós foram verificadas na prévia local.

### 2026-07-28 — Kit & Árvore / Forte Meridian

- O mapa radial anterior foi substituído pela composição vertical escolhida pelo usuário.
- A tela segue `skillTree`, `parentNodes`, `branchIds` e `skills`; campos ausentes
  mostram `Não disponível`.
- Adicionado fundo atmosférico raster dedicado e modo imersivo que recolhe o hero
  do personagem apenas na aba Kit & Árvore.
- Seleção de nós, alternância de modos e busca foram validadas no browser sem erros.

### 2026-07-28 — Kit & Árvore / Resonance Lens

- Substituída a grade técnica pelo mapa radial selecionado na ideação visual.
- Modos, nós, dependências, níveis, valores e guia Forte usam somente dados da API.
- Adicionada uma fixture DEV reduzida de Aemeath baseada no banco 3.6.1 para QA visual.
- Build frontend e interações principais validados sem erros.

### 2026-07-28 — fullscreen e memória para IA

- Builds passou a usar o mesmo shell sem largura máxima de Equipes; as colunas
  laterais têm limites responsivos e o conteúdo central ocupa o espaço disponível.
- O espaçamento interno da tela de Builds usa um gutter fluido entre 16 e 24 px,
  compartilhado por cabeçalho, compositor, inspetor, filtros e biblioteca.
- A memória operacional foi consolidada em `prompt/`. `prompt/start.md` é o ponto
  de entrada, `architecture.md` registra a estrutura e `brain.md` mantém o estado.
- Capturas e comparações usadas apenas durante QA visual não fazem parte do produto.
- `prompt/` foi reduzida aos três arquivos essenciais: `start.md`,
  `architecture.md` e `brain.md`.
- O `README.md` principal foi refeito para documentar somente o produto atual,
  sua estrutura, execução, validação, build, dados locais e continuidade com IA.
- O teste do workspace deixou de acessar `DashboardSummary.Goals`, removido junto
  com o Planejador; metas e convenes são validadas por seus repositórios próprios.

### 2026-07-28 — brain para IA

- Mantidos somente `start.md`, `architecture.md` e `brain.md`.
- Definido o protocolo de leitura e atualização obrigatória para futuras IAs.
- Registrada a arquitetura atual, decisões duráveis, riscos e próximos trabalhos.

## Como atualizar este arquivo

Ao terminar uma tarefa, altere apenas o necessário:

1. Atualize “Estado implementado” se uma capacidade mudou.
2. Adicione ou remova riscos reais.
3. Atualize próximos trabalhos quando algo for concluído ou descoberto.
4. Acrescente uma entrada curta em “Registro recente”.
5. Remova entradas antigas que já não ajudam a próxima IA.
