package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	Token, err := load_token()
	log_err("Error Loading Token", err)
	dg, err := discordgo.New("Bot " + Token)
	log_err("Error creating Discord Session", err)

	//register callbacks
	dg.AddHandler(ready)
	//open the websocket
	err = dg.Open()
	log_err("Error opening Discord Session", err)
	fmt.Println("Running . . .")
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-signal_channel
	//Gracefully Shutdown
	dg.UpdateGameStatus(1, "")

	dg.Close()
	fmt.Println("Stopping . . .")
}
func ready(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateGameStatus(0, "Google Classroom…")
	session.UpdateListeningStatus("Mr. Oase")
}
func load_token() (string, error) {
	// load environment variables from .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading .env file")
		return "", err
	}
	Token := os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		return "", errors.New("no token loaded")
	} else {
		return Token, nil
	}
}
func log_err(msg string, err error) {
	if err != nil {
		fmt.Println(msg)
		log.Println(err)
	}
}
