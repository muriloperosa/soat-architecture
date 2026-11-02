# Roteiro de apresentação dos endpoints da Fase 2

Este roteiro acompanha a collection .dev/postman/soat-fase-2.postman_collection.json. As falas entre aspas podem ser usadas durante a gravação; as linhas em itálico indicam a chamada a executar no Postman.

> A collection está configurada para o ambiente publicado. As chamadas criam registros e podem enviar e-mails reais.

### Configurar o e-mail de destino no Postman

Antes de executar a collection, abra a aba **Variables** dela e substitua `seu-email@exemplo.com` em `cliente_email_base` pelo endereço que deve receber os e-mails. Salve a collection. O script acrescenta `+f2` e um identificador ao endereço em cada execução; no Gmail, essa variação chega à mesma caixa. Outros provedores precisam aceitar endereços com `+`.

Não altere os bodies das requests para configurar o destinatário. O envio real depende de o SMTP já estar configurado no ambiente publicado.

## 1. Abertura

```text
Vou demonstrar o fluxo principal da API da oficina: autenticação, abertura e acompanhamento de uma Ordem de Serviço, preparação do orçamento e decisão do cliente. Também vou mostrar a listagem de OS com e sem filtro de status.

Estou usando a collection da Fase 2 no Postman. Os cadastros iniciais preparam os dados necessários para executar o fluxo.
```

## 2. Login e preparação dos dados

*Execute `01 - Login admin`.*

```text
A primeira chamada autentica um usuário interno. A collection guarda o token retornado e o utiliza nas operações administrativas seguintes.
```

*Execute `02 - Criar cliente` e `03 - Login cliente`.*

```text
Agora crio um cliente para este cenário e faço login com ele. A decisão do orçamento será feita usando o token do próprio cliente.
```

*Execute `04 - Cadastrar veículo` e `05 - Cadastrar serviço`.*

```text
Também cadastro um veículo e um serviço. Esses dados são necessários para abrir a OS e montar o orçamento. A collection guarda os identificadores retornados para usá-los nas próximas chamadas.
```

## 3. Abertura e consulta da OS

*Execute `06 - Abrir OS` e `07 - Consultar status da OS`.*

```text
A OS é criada no status RECEBIDA. Em seguida, consulto a OS pelo identificador para conferir o status inicial.
```

*Execute `08 - Listar OS sem filtro de status por prioridade`.*

```text
Sem informar um status, a API ordena as OS pela prioridade definida pela aplicação e, dentro de cada prioridade, mostra primeiro as mais antigas. A listagem padrão também exclui OS finalizadas e entregues.
```

*Execute `09 - Listar OS com filtro de status RECEBIDA`.*

```text
Agora filtro a listagem por RECEBIDA. A resposta contém somente OS nesse status e mantém a ordenação por antiguidade dentro do grupo.
```

## 4. Diagnóstico e orçamento

*Execute `10 - Iniciar diagnóstico` e `11 - Registrar diagnóstico`.*

```text
A OS passa para EM_DIAGNOSTICO, e registro o resultado da avaliação.
```

*Execute `12 - Gerar orçamento` e `13 - Adicionar serviço ao orçamento`.*

```text
Com o diagnóstico registrado, crio o orçamento e adiciono o serviço para compor o valor que será apresentado ao cliente.
```

*Execute `14 - Enviar orçamento para aprovação` e `15 - Confirmar status AGUARDANDO_APROVACAO`.*

```text
Ao enviar o orçamento, a OS muda para AGUARDANDO_APROVACAO e a aplicação aciona a notificação por e-mail. Em seguida, consulto a OS para confirmar a mudança de status.
```

## 5. Decisão do cliente

*Execute `16 - Aprovar ou rejeitar orçamento (cliente)`.*

```text
A decisão usa o token do cliente. Na variável decisao_orcamento, posso escolher aprovar ou rejeitar. Neste exemplo, vou aprovar o orçamento.
```

*Execute `17 - Consultar status após decisão`.*

```text
A consulta confirma que o status mudou para APROVADA. Se eu tivesse escolhido rejeitar, o status seria REJEITADA.
```

## 6. Execução e entrega

*Execute `18 - Iniciar execução (somente se aprovada)`, `19 - Finalizar OS (somente se aprovada)` e `20 - Entregar OS (somente se aprovada)`.*

```text
Como o orçamento foi aprovado, a equipe inicia a execução, finaliza o serviço e registra a entrega. Essas mudanças de status também podem gerar notificações por e-mail.
```

*Execute `21 - Verificar filtro de OS encerradas`.*

```text
Por fim, confirmo que a OS entregue não aparece na listagem padrão, que exclui os status FINALIZADA e ENTREGUE.
```

## 7. Encerramento

```text
Esse fluxo demonstra os endpoints principais da Fase 2: autenticação, abertura e consulta de OS, listagem priorizada com e sem filtro, diagnóstico, orçamento e decisão do cliente. Para demonstrar a rejeição, basta alterar decisao_orcamento para rejeitar e executar novamente a collection.
```
