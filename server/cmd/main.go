package main

import (
	"context"
	"flag"

	botApp "github.com/Guise322/TeleBot/server/internal/app/bot"
	"github.com/Guise322/TeleBot/server/internal/app/interruption"
	botEnt "github.com/Guise322/TeleBot/server/internal/entities/bot"
	botInfra "github.com/Guise322/TeleBot/server/internal/infrastructure/bot"
	"github.com/Guise322/TeleBot/server/internal/infrastructure/config"
	serviceInfra "github.com/Guise322/TeleBot/server/internal/infrastructure/service"

	"github.com/sirupsen/logrus"
)

func main() {
	configPath := flag.String("conf", "../etc/config.yml", "Config path.")
	flag.Parse()

	appCtx, appCancel := context.WithCancel(context.Background())

	logrus.Infoln("Start the application")

	interruption.WatchForInterruption(appCancel)

	configer := config.NewConfiger(*configPath)
	config, err := configer.LoadConfig()
	if err != nil {
		logrus.Errorf("Can not get the config data: %v", err)

		return
	}

	commands := &[]botEnt.Command{}
	bot, err := botInfra.NewTelebot(config.BotKey, commands)
	if err != nil {
		logrus.Errorf("A bot error occurs: %v", err)

		return
	}

	receiver := serviceInfra.NewKafkaReceiver(config.KafkaAddress)
	transmitter := serviceInfra.NewKafkaTransmitter(config.KafkaAddress)

	err = botApp.Process(appCtx, bot, receiver, transmitter, commands)
	if err != nil {
		logrus.Errorf("An error occurs: %v", err)

		return
	}
}
