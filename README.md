# TelegramNotify

Este projeto foi criado inicialmente para treinar Go, integrando o Zabbix com a API de Bots do Telegram, apenas para verificar os dispositivos online. Porém, diante da praticidade para mim, foram incrementadas novas funcionalidades.

Por se tratar de algo específico para mim, provavelmente as consultas não funcionarão para todos, sendo necessários ajustes no código, visto que não generalizei muito as consultas.

> ⚠️ **Nota**: Este é um projeto pessoal e não foi desenvolvido para ser aplicado de forma genérica. Você precisará adaptar as consultas ao Zabbix, IDs de grupos, nomes de items e outras configurações específicas do seu ambiente.

Mas, caso alguém queira tentar, você pode começar fazendo o seguinte.

## 🔧 Passos Iniciais

Utilize o `.env.example` para definir as variáveis abaixo:

```dotenv
TELEGRAM_API_TOKEN=<YOUR_TELEGRAM_BOT_TOKEN>
TELEGRAM_ALLOWED_CHAT_ID=123,456
ZABBIX_API_TOKEN=<YOUR_ZABBIX_TOKEN>
ZABBIX_API_URL=<YOUR_ZABBIX_SERVER_ADDRESS>/zabbix/api_jsonrpc.php
```

### Configuração do Telegram

Para o Telegram, você precisará chamar o [@BotFather](https://t.me/botfather) e criar uma chave. Como o Telegram não disponibiliza uma ferramenta de visibilidade do bot, foi necessário fazer a validação via código, onde eu capturo o ChatID (aparecerá no log assim que seu bot for acionado) e defino que ele está autorizado a conversar com esse Chat.

#### <span style="color:red">⚠️ ATENÇÃO!</span>

Cuide bem da sua chave, pois, qualquer um com acesso a ela, terá controle total de seu Bot.

### Configuração do Zabbix

Para o Zabbix, existem duas alternativas: você capturar o token via autenticação, ou definir um token já no Zabbix. Por praticidade e facilidade de revogação caso necessário, optei pela segunda opção.

Será necessário definir o TOKEN e o endereço do seu servidor Zabbix, lembrando de manter o `/api_jsonrpc.php` que é o ponto de chamada da API.

**Documentação oficial:**
- [Telegram Bot API](https://core.telegram.org/bots/tutorial#introduction)
- [Zabbix API](https://www.zabbix.com/documentation/current/en/manual/api)

### ⚙️ Ajustes Necessários

Como este projeto foi desenvolvido para um ambiente específico, você precisará ajustar:

- **IDs de grupos no Zabbix**: No arquivo `printers.go`, o grupo de impressoras está hardcoded como `"22"`. Ajuste para o ID do seu grupo.
- **Keys de items**: Os items buscados (como `"icmpping"`, `"contador.colorido"`, `"TOTVS"`) precisam existir no seu Zabbix com os mesmos nomes, ou você deve alterar o código.
- **Comandos Windows**: Os comandos de restart/shutdown funcionam apenas em ambientes Windows com permissões adequadas.

## 🚀 Funcionalidades

### 🌐 Monitoramento de Rede

#### `/ping <ip1> <ip2> ...`
Realiza ping em um ou mais endereços IP simultaneamente.
- Mostra latência média, pacotes enviados/recebidos e taxa de perda
- Suporta múltiplos IPs em uma única execução
- Exemplo: `/ping 192.168.0.1 192.168.0.2`

#### `/listip`
Lista todos os hosts e seus endereços IP cadastrados no Zabbix.
- Exibe hostname e interface(s) de rede
- Apenas hosts ativos são listados

### 📊 Monitoramento Zabbix

#### `/status_check`
Verifica o status online/offline de todos os hosts monitorados.
- ✅ Host online (icmpping = 1)
- ❌ Host offline (icmpping = 0)
- Consulta em tempo real via API Zabbix

#### `/printers_counter`
Exibe os contadores de impressão das impressoras monitoradas.
- Contador preto e branco
- Contador colorido
- Contador total
- Apenas impressoras do grupo específico no Zabbix (ID: 22)

#### `/protheus_status`
Monitora o status dos serviços Protheus/TOTVS.
- ✅ Serviço rodando
- ❌ Serviço parado
- Consulta itens com key "TOTVS" no Zabbix

### 💻 Gerenciamento de Hosts Windows

#### `/restart_win <hostname>`
Reinicia remotamente um host Windows.
- Requer permissões administrativas
- Exemplo: `/restart_win LVMAQUINA`

#### `/shutdown_win <hostname>`
Desliga remotamente um host Windows.
- Requer permissões administrativas
- Exemplo: `/shutdown_win LVMAQUINA`

### ⏰ Sistema de Agendamento

#### `/schedule_add <min> <hora> <dia> <mês> <dia_semana> <comando>`
Cria um novo agendamento usando expressões CRON.
- Executa comandos automaticamente no horário especificado
- Suporta todos os comandos do bot
- Exemplo: `/schedule_add 0 8 20 * * printers_counter`
    - Executa `/printers_counter` todo dia 20 às 08:00

#### `/schedule_list`
Lista todos os agendamentos ativos.
- Mostra ID, expressão CRON e comando de cada agendamento
- IDs são necessários para remover agendamentos

#### `/schedule_remove <ID>`
Remove um agendamento específico pelo ID.
- Exemplo: `/schedule_remove 1764686892095287000`

#### `/schedule_help`
Exibe guia completo sobre expressões CRON.
- Exemplos práticos de agendamentos
- Formato: `MIN HORA DIA MÊS DIA_SEMANA`

## 📦 Dependências

```bash
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get github.com/joho/godotenv
go get github.com/go-ping/ping
go get github.com/go-co-op/gocron
```

### Desenvolvido com o apoio destes pacotes

```
github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
github.com/go-co-op/gocron v2.x.x
github.com/joho/godotenv v1.5.1
github.com/go-ping/ping v1.2.0
```

## 🚀 Execução

```bash
go run main.go
```

## 🔐 Segurança

- Apenas chat IDs autorizados podem usar o bot
- Comandos Windows requerem privilégios administrativos
- Agendamentos são persistidos em `schedules.json`
- Nunca compartilhe seu token do Telegram ou do Zabbix

## 📝 Logs

Todas as interações são registradas no console com:
- Chat ID do usuário
- Nome e username
- Comando executado
- Timestamp