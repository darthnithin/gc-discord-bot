package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"google.golang.org/api/classroom/v1"
)

var (
	Token    string
	Class_id string
	err      error
	srv      *classroom.Service
)

/* var srv = Class(Class_id) */

func main() {
	Token, Class_id, err = load_token()
	log_err("Error Loading Token", err)

	srv = Class(Class_id)
	if srv == nil {
		fmt.Println("niiiil")
	}
	dg, err := discordgo.New("Bot " + Token)
	log_err("Error creating Discord Session", err)
	//load classroom stuff
	/* 	s, err := announce(Class_id, srv)
	   	log_err("loading classroom...", err)
	   	for _, announcement := range s.Announcements {
	   		println(announcement.Text)
	   	} */
	//register callbacks
	dg.AddHandler(ready)
	dg.AddHandler(messageCreate)
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
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {

	PREFIX := "!classroom"
	content := message.Content

	fmt.Println(content)
	if message.Author.ID == session.State.User.ID {
		return
	}
	if len(content) <= len(PREFIX) {
		return
	}
	if content[:len(PREFIX)] != PREFIX {
		return
	}
	content = content[len(PREFIX):]
	args := strings.Fields(content)
	channel, err := session.State.Channel(message.ChannelID)
	log_err("Could not find channel", err)
	// find the guild (server) of the message
	_, err = session.State.Guild(channel.GuildID)
	log_err("Could not find guild", err)
	if args[0] == "list" {
		if args[1] == "classes" || args[1] == "courses" {
			CourseList := list_courses(srv, 10)
			var ReplyText string
			for _, class := range CourseList {
				ReplyText += class.Name
				ReplyText += "\n"
				//m, err := json.Marshal(class)
				fmt.Printf("| %v", class)
				//fmt.Printf("marshalled %v", string(m))
				//log_err("err marrshalling", err)
			}
			_, err := session.ChannelMessageSendReply(channel.ID, ReplyText, message.Reference())
			log_err("Err Replying", err)

		}
		if args[1] == "announcements" {

		}
		/* 		s, err := announce(Class_id, srv)
		   		log_err("error announce", err)
		   		for _, announcement := range s.Announcements {
		   			_, _ = session.ChannelMessageSend(channel.ID, announcement.Text)
		   		} */

	}
	fmt.Println(args)
	// find the channel of the message

	// send a message
	log_err("error sending message", err)

}
func ready(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateGameStatus(0, "Google Classroom…")
	session.UpdateListeningStatus("Listening to \"!classroom\" ")
}
func load_token() (string, string, error) {
	// load environment variables from .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("error loading .env file")
		return "", "", err
	}
	Token := os.Getenv("DISCORD_TOKEN")
	Class_id := os.Getenv("CLASS_ID")
	if Token == "" || Class_id == "" {
		return "", "", errors.New("token (or class id) not loaded")
	} else {
		return Token, Class_id, nil
	}

}
func log_err(msg string, err error) {
	if err != nil {
		fmt.Println(msg)
		log.Println(err)
	}
}
