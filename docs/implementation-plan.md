# Plano de implementação

## Fase 0 — concluída

- Inspeção integral de `prompt/start.md`, `prompt/architecture.md`,
  `prompt/brain.md` e `wuwa_scraper.py`.
- Verificação de Go, Node.js, npm e Wails.
- Registro da arquitetura e dos comportamentos reutilizáveis.
- Validação do manifest e do schema real do índice do Nanoka.

## Fase 1 — fundação funcional

- [x] Shell Wails, React e TypeScript estrito.
- [x] Design tokens e primeira tela do catálogo.
- [x] SQLite, WAL, migrations incorporadas e FTS5.
- [x] Separação entre domínio, casos de uso, repositório e fonte externa.
- [x] Preferências persistidas de sidebar, densidade, movimento e IA.
- [x] Filtros e modo grade/tabela persistidos no frontend.
- [x] Navegação funcional, breadcrumbs, sidebar recolhível e atalhos Ctrl+K/Ctrl+B/Ctrl+,/Escape.
- [ ] Ação contextual de Ctrl+N.
- [ ] Logs em arquivo com rotação e tela de diagnóstico.

## Fase 2 — sincronização

- [x] Detecção de versão e índice de personagens.
- [x] Retry transitório, timeout, cancelamento e limite de resposta.
- [x] Upsert transacional sem sobrescrever dados pessoais.
- [x] Snapshot consistente antes da sincronização e rollback com cópia de segurança.
- [ ] ETag, Last-Modified e cache HTTP.
- [x] Detalhes de personagens, skills, chains e armas assinatura.
- [x] Download, validação e cache atômico de imagens WebP.
- [x] Progresso detalhado e cancelamento pela interface.

## Fase 3 — catálogo

- [x] Grade, tabela, busca e filtros essenciais.
- [x] Ordenação original do catálogo da API preservada no SQLite.
- [x] Estados vazio, carregando e erro.
- [ ] Virtualização para listas extensas.
- [x] Posse, nível, sequência e favoritos editáveis.
- [x] Página de personagem com visão geral, kit e sequências.
- [x] Materiais oficiais de ascensão, Forte e multiplicadores por nível.
- [x] Planejador de nível/habilidades com totais calculados no domínio Go.
- [x] Tags de função, guia do Forte, lore, atributos, fraqueza e árvore completa.
- [x] Catálogo, página detalhada e inventário local de armas.

## Fases seguintes

## Fase 4 — builds e equipes funcional

- [x] CRUD de builds com personagem e arma.
- [x] Duplicação, favorito, bloqueio e exclusão com undo.
- [x] Versão do jogo, níveis, sequência, refinamento e notas.
- [x] Cinco Echoes, main stats e substats.
- [x] Buffs externos, condições, inimigo e rotação associada.
- [x] Histórico automático de versões.
- [ ] Importação/exportação.

## Fase 5 — IA funcional

- [x] Ollama e LM Studio locais.
- [x] Gemini remoto com chave somente em memória.
- [x] Modos estrito, assistido e geral.
- [x] Chat contextual de builds e equipes com histórico.
- [x] Streaming, teste de conexão, descoberta de modelos e RAG estruturado por FTS5.
- [x] Contexto de personagem e guias oficiais sincronizados sob demanda.
- [ ] Embeddings e ferramentas com gravação confirmável.

## Fases 6–8 — núcleo funcional

- [x] Calculadora determinística, breakdown, inimigos e fórmula versionada.
- [x] Echoes, rotações sequenciais e DPS.
- [x] Dashboard, conta, planejador e Convene Tracker manual.
- [x] Timeline avançada, cooldowns, energia, buffs temporais e avisos.
- [x] Comparador de builds/equipes com DPS e exportação JSON/Markdown/PNG/PDF.
- [x] Backup completo, restauração e importação/exportação portátil validada.
- [x] Materiais oficiais.
- [ ] Automatização Playwright do aplicativo empacotado.

## Bloqueios e incertezas reais

- Ascensões, custos de skills, stats e multiplicadores foram validados com o
  payload real `character/1606.json` da versão `3.6.1`; os DTOs continuam
  tolerantes a campos adicionais.
- A referência visual “Healer Comparison” não está presente na pasta; o
  comparador não deve ser finalizado sem esse asset.
- A política exata de cache e retenção de snapshots precisa ser definida depois
  de medir o tamanho real do conjunto completo de dados e imagens.
- Bancos legados sem `schema_migrations` são reconciliados automaticamente desde
  a migration `0005`; manter testes desse caminho em todas as próximas migrations.
