# 0010. SMTP no lugar do Amazon SES no AWS Academy Learner Lab

## Status
Aceita.

## Contexto
A notificação de mudança de status da ordem de serviço e o envio do orçamento com os links de aprovação dependem de e-mail. O provedor escolhido para a solução é o **Amazon SES**: é gerenciado, roda na mesma conta do cluster e autentica pela role dos nós, sem senha no pod. O adapter `ses` foi implementado e testado.

O ambiente roda no AWS Academy Learner Lab. O SES só envia a partir de identidades verificadas, e o lab nega `ses:VerifyEmailIdentity`. Também não é possível tirar a conta do sandbox. Sem remetente verificado, nenhuma mensagem sai.

## Decisão
- **SES continua sendo o provedor-alvo.** O adapter `ses` fica no código, e a escolha é feita por `EMAIL_PROVIDER`. Numa conta AWS sem a restrição do lab, basta `EMAIL_PROVIDER=ses` e um `SES_FROM_ADDRESS` verificado, sem mudança de código.
- **No Learner Lab, o envio usa SMTP com STARTTLS** (`EMAIL_PROVIDER=smtp`), com uma conta Gmail e senha de app.
- **Senha fora do repositório e do state:** a senha de app entra como secret `SMTP_PASSWORD` do environment `aws-deploy`. O Terraform cria o secret `oficina/smtp` no Secrets Manager com o argumento write-only `secret_string_wo`, e o CD copia o conteúdo para o Secret Kubernetes `oficina-api-secrets`.
- **SMTP obrigatório no CD:** o Terraform exige `SMTP_PASSWORD` e sempre cria `oficina/smtp`, e o CD configura `EMAIL_PROVIDER=smtp` a partir dele. O ramo do CD que escolhe `ses` quando há `SES_FROM_ADDRESS` em `oficina/api` fica sem uso no lab.

Alternativas descartadas:

- **Usar o SES no sandbox com identidades verificadas pelo console:** o lab também bloqueia a verificação pelo console.
- **Outro provedor transacional (SendGrid, Mailgun etc.):** exigiria conta, chave de API e adapter novos só para o ambiente acadêmico; o SMTP do Gmail resolve com o mesmo protocolo padrão.
- **Deixar `EMAIL_PROVIDER=log` no EKS:** o requisito de aviso por e-mail não seria demonstrável.

## Consequências
Positivas:
- O requisito de e-mail funciona no ambiente do lab.
- Na aplicação, voltar ao SES é só configuração. No pipeline, exige tornar o SMTP opcional no Terraform e no CD, verificar o remetente no SES e garantir `ses:SendEmail` na role dos nós.

Negativas:
- O ambiente depende de uma conta Gmail e de uma senha de app, que precisa ser rotacionada à mão (secret `SMTP_PASSWORD` mais `SMTP_SECRET_VERSION`).
- O Gmail limita o volume diário de envio, adequado só para demonstração.
- A saída na porta 587 precisa estar liberada na rede do lab.
