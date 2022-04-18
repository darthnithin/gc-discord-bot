package main

import (
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

var srv *classroom.Service

func main() {
	// Load Token
	Token, Class_id := load_token()
	// Load Google Classroom Service
	srv = Class(Class_id)
	// Load DiscordGo Session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Println("Error Registering discordgo Session. Error: %v", err)
	}
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
	log_err("Could not find channel", err, false)
	// find the guild (server) of the message
	_, err = session.State.Guild(channel.GuildID)
	log_err("Could not find guild", err, false)
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
			log_err("Err Replying", err, false)

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
}
func ready(session *discordgo.Session, event *discordgo.Ready) {
	session.UpdateGameStatus(0, "Google Classroom…")
	session.UpdateListeningStatus("Listening to \"!classroom\" ")
}
func load_token() (string, string) {
	// load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Panicf("error loading .env file:%v", err)
		return "", ""
	}
	Token := os.Getenv("DISCORD_TOKEN")
	Class_id := os.Getenv("CLASS_ID")
	if Token == "" || Class_id == "" {
		log.Panicf("Token (or class id) not loaded. Make sure your .env file includes your Discord Token & Class_ID")
		return "", ""
	} else {
		return Token, Class_id
	}

}
