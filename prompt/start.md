# START — WaveArchive

Este é o ponto de entrada para qualquer IA que vá modificar o WaveArchive.

## Protocolo obrigatório

Antes de alterar código:

1. Leia `prompt/architecture.md` por completo.
2. Leia `prompt/brain.md` por completo.
3. Leia o `README.md` principal e a documentação relevante em `docs/`.
4. Verifique o estado atual do repositório e preserve alterações do usuário.
5. Confirme no código a arquitetura descrita; o código é a fonte final da verdade.

Durante o trabalho:

1. Respeite a separação `fonte externa → mapper → domínio → caso de uso →
   repositório → binding Wails → React`.
2. Não exponha DTOs externos diretamente ao frontend.
3. Não invente personagens, tags, materiais, builds, equipes ou resultados.
4. Use dados da API, cálculos determinísticos ou dados inseridos pelo usuário.
5. Não registre chaves, tokens, prompts privados ou dados sensíveis.
6. Preserve a operação local-first e a compatibilidade com bancos existentes.
7. Crie migrations novas; nunca edite uma migration já aplicada.

Antes de encerrar:

1. Atualize `prompt/brain.md` com o que realmente foi alterado.
2. Atualize `prompt/architecture.md` somente se a estrutura, fluxo de dados ou
   contrato entre camadas mudou.
3. Registre bloqueios reais e decisões pendentes, sem esconder falhas.
4. Informe os arquivos modificados e quais validações foram ou não executadas.

## Ordem de autoridade

Quando houver conflito, siga esta ordem:

1. Pedido atual do usuário.
2. Código e contratos atuais.
3. `prompt/architecture.md`.
4. `prompt/brain.md`.
5. `README.md` e `docs/`.

O brain é memória operacional, não autorização para ampliar o escopo.
