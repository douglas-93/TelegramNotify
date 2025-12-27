# Comando: `/list_services`

## Descrição

Lista todos os serviços de um host Windows remoto, com opção de filtrar por nome.

## Uso

```bash
/list_services <IP/Host> [filtro]
```

### Parâmetros

- **`<IP/Host>`** (obrigatório): Endereço IP ou nome do host remoto
- **`[filtro]`** (opcional): String para filtrar serviços por nome (case-insensitive)

## Exemplos

### Listar todos os serviços

```bash
/list_services 192.168.100.16
```

**Resultado:**

```
📋 Serviços em 192.168.100.16

Total: 245 serviços

• Adobe Acrobat Update Service
• Application Experience
• Application Information
• AppX Deployment Service (AppXSVC)
...
(mostra até 50 serviços)

... e mais 195 serviços

💡 Dica: Use um filtro para refinar a busca
```

### Listar serviços filtrados

```bash
/list_services 192.168.100.16 TOTVS
```

**Resultado:**

```
🔍 Serviços em 192.168.100.16 (filtro: totvs)

Total: 8 serviços

• TOTVS_AppServer_P12
• TOTVS_AppServer_P12_Balance
• TOTVS_DBAccess_P12
• TOTVS_LicenseServer
• TOTVS_Logix_AppServer
• TOTVS_Logix_DBAccess
• TOTVS_Protheus_REST
• TOTVS_Protheus_WebAgent
```

## Funcionalidades

### ✅ Feedback em Tempo Real

O comando fornece feedback progressivo:

1. **⏳ Conectando em [host]...**
2. **📋 Listando serviços...**
3. **Resultado final**

### ✅ Limitação de Exibição

- Mostra até **50 serviços** por vez
- Se houver mais de 50, exibe contador e sugere usar filtro
- Evita mensagens muito grandes no Telegram

### ✅ Filtro Inteligente

- Busca **case-insensitive** (não diferencia maiúsculas/minúsculas)
- Busca por **substring** (encontra "totvs" em "TOTVS_AppServer")
- Funciona com qualquer parte do nome do serviço

### ✅ Formatação Markdown

- Usa formatação Markdown para melhor legibilidade
- Serviços exibidos em código inline (`` ` ``)
- Destaque em negrito para informações importantes

## Casos de Uso

### 1. Descobrir Serviços Disponíveis

```bash
/list_services SERVER01
```

### 2. Encontrar Serviços Específicos

```bash
/list_services SERVER01 SQL
/list_services SERVER01 Apache
/list_services SERVER01 Windows
```

### 3. Verificar Serviços de Aplicação

```bash
/list_services 192.168.1.100 Protheus
/list_services 192.168.1.100 TOTVS
/list_services 192.168.1.100 SAP
```

### 4. Preparar para Gerenciar Serviços

Use `/list_services` para descobrir o nome exato do serviço, depois use `/services` para gerenciá-lo:

```bash
# 1. Listar serviços
/list_services SERVER01 TOTVS

# 2. Gerenciar serviço específico
/services SERVER01 restart TOTVS_AppServer_P12
```

## Requisitos

- **Permissões**: O bot deve ter permissões administrativas no host remoto
- **Firewall**: RPC deve estar permitido entre o bot e o host remoto
- **Rede**: Conectividade de rede com o host

## Tratamento de Erros

### Erro de Conexão

```
❌ Erro ao conectar no host 192.168.100.16: RPC server is unavailable
```

**Possíveis causas:**

- Host offline
- Firewall bloqueando RPC
- Credenciais insuficientes

### Erro ao Listar

```
❌ Erro ao listar serviços: Access is denied
```

**Possíveis causas:**

- Permissões insuficientes
- Serviço SCM não disponível

## Integração com Outros Comandos

### Fluxo de Trabalho Completo

1. **Descobrir serviços:**

   ```bash
   /list_services SERVER01 TOTVS
   ```

2. **Gerenciar serviço:**

   ```bash
   /services SERVER01 restart TOTVS_AppServer_P12
   ```

3. **Verificar conectividade:**

   ```bash
   /ping 192.168.100.16
   ```

## Dicas

💡 **Use filtros específicos** para encontrar rapidamente o que procura

💡 **Copie o nome exato** do serviço para usar com `/services`

💡 **Teste a conexão** com `/ping` antes de listar serviços

💡 **Verifique permissões** se receber erro de acesso negado
