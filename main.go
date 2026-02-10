package main

import (
	"LapaTelegramBot/config"
	bot "LapaTelegramBot/telegram"
)

func main() {
	config.Load()
	bot.StartBot()
}
