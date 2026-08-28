# Segurança: ferramentas de análise de vulnerabilidades

Este documento registra as ferramentas usadas para cobrir as três camadas clássicas de segurança (SCA, SAST, DAST) no projeto, complementando o SonarQube Community, que tem regras de segurança limitadas nesse plano.

## Objetivo

Produzir um relatório de vulnerabilidades com evidência real de três frentes independentes: dependências (SCA), código fonte próprio (SAST) e API em execução (DAST). Cada ferramenta cobre uma superfície de ataque distinta; nenhuma sozinha seria suficiente.

## SCA: govulncheck

Ferramenta oficial da equipe Go, mantida em `golang.org/x/vuln`.

Justificativa: detecta vulnerabilidades conhecidas (CVE/GHSA) tanto em dependências do `go.mod` quanto na stdlib. O diferencial em relação a scanners genéricos de SCA é a análise de alcançabilidade: o `govulncheck` só alerta quando o binário realmente chama a função vulnerável, não apenas quando a dependência vulnerável está no grafo de imports. Isso reduz drasticamente falso positivo, já que a maioria dos scanners de dependência aponta CVE em código morto ou em caminho nunca exercitado.

Instalação:
```
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Execução:
```
govulncheck -json ./...
```

## SAST: gosec

Justificativa: gosec é nativo de Go, analisa a AST do próprio código fonte (não bytecode, não regra genérica). Cobre padrões de risco específicos da linguagem: uso de `math/rand` em contexto de segurança, SQL injection via concatenação de string, permissão de arquivo insegura, uso de `crypto/md5`/`sha1` em contexto sensível, hardcoded credentials, `G401`-`G404` (crypto fraco), entre outras regras mapeadas por CWE. Complementa o SonarQube Community porque as regras de segurança do plano free do Sonar são rasas para Go; gosec cobre esse gap sem custo de licença.

Instalação:
```
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

Execução:
```
gosec -fmt=json -out=gosec-report.json ./...
```

## DAST: OWASP ZAP

Justificativa: SCA e SAST analisam código estático; nenhum dos dois prova que a API, rodando de verdade, resiste a ataque. ZAP simula ataques reais contra a API em execução (SQLi, broken auth, injection, etc). O modo API scan aceita `swagger.json`/`openapi.yaml` como entrada, mapeando e testando todos os endpoints documentados automaticamente, sem precisar de script de navegação manual. Gera relatório em HTML/PDF, adequado como apêndice.

Pré-requisito: API rodando contra dados de teste (nunca produção). Se a rota exige JWT, configurar script de autenticação no ZAP, senão o scan só alcança endpoint público.

### Execução: `make sec-dast`

`security/zap-scan.sh` orquestra o scan e sobe uma stack Docker isolada, definida em `compose.tools.yml` (serviços `app-dast` + `mysql-dast`, projeto compose separado do dev via `-p soat-architecture-dast`). Motivo: o scan autenticado escreve dados reais (ordens de serviço, veículos, peças) via requisições ativas do ZAP, então rodar contra o `mysql`/`app` de desenvolvimento (`compose.yml`) sujaria o banco local. `mysql-dast` usa `tmpfs`, e a stack inteira é derrubada (`down -v`) ao fim do script — nada persiste, o banco de dev nunca é tocado.

O script roda o scan autenticado **uma vez por papel** (`admin`, `atendente`, `mecanico`, `cliente`), cada um com seu próprio token JWT injetado via `replacer` do ZAP. Isso não é só cobertura de rota: prova que RBAC bloqueia de fato um papel tentando acessar rota de outro (autorização negativa), não só que as rotas respondem quando autenticado. Relatórios saem em `security/reports/zap-report-{admin,atendente,mecanico,cliente}.{html,json}`.

### Falsos positivos conhecidos (`security/zap-rules.tsv`)

Três alertas disparam sempre em `/swagger/doc.json` e `/swagger` e são falsos positivos confirmados (não são erro real da API):

- `90022` (Application Error Disclosure) e `10023` (Debug Error Messages): a string "Internal Server Error" que o ZAP encontra é a *descrição* de uma resposta 500 documentada no próprio OpenAPI spec (`docs/swagger/swagger.json`), não um erro de fato — confirmado sem nenhuma resposta `500` real nos logs da app durante o scan.
- `100001` (Unexpected Content-Type): `/swagger` é a página HTML do Swagger UI, não um endpoint de API — `text/html` é o content-type correto ali.

`zap-rules.tsv` marca esses três `pluginid`s como `IGNORE` (usado via `-c zap-rules.tsv` no serviço `zap`), o que evita que virem `FAIL` no veredito do ZAP. Isso sozinho não remove os alertas dos relatórios HTML/JSON — por isso `zap-scan.sh` também filtra esses `pluginid`s do JSON gerado (função `strip_ignored_alerts`) depois de cada scan.