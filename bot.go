package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"google.golang.org/api/classroom/v1"
)

var srv *classroom.Service
var waitingforuser bool
var result = make(chan *classroom.Service, 1)
var ids = make(chan string, 1)
var dms = make(chan string, 1)
var tokenstream = make(chan string, 1)

func main() {
	// Load Token
	Token := load_token()
	// Load Google Classroom Service

	go webserver(tokenstream)
	// Load DiscordGo Session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Printf("Error Registering discordgo Session. Error: %v", err)
	}
	waitingforuser = false
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
	if message.Author.ID == session.State.User.ID {
		return
	}
	if waitingforuser && (message.ChannelID == <-ids) {
		dms <- content
	}
	if len(content) <= len(PREFIX) {
		return
	}
	if content[:len(PREFIX)] != PREFIX {
		return
	}
	log.Printf("command detected: %v", content)
	content = content[len(PREFIX):]
	args := strings.Fields(content)
	channel, err := session.State.Channel(message.ChannelID)
	//logging
	if err != nil {
		log.Printf("%v", err)
	}
	// This is so i don't have to manually pass every single required parameter down the functions.
	info := discord_info{User: message.Author, Session: session, channel: channel, msg_ref: message.Reference(), ids: ids}
	go Class(&info, result)
	srv = <-result
	if args[0] == "list" {
		if args[1] == "classes" || args[1] == "courses" {

			CourseList := list_courses(srv, 10)
			var ReplyText string
			for _, class := range CourseList {
				ReplyText += class.Name
				ReplyText += "\n"
				//m, err := json.Marshal(class)
				//fmt.Printf("| %v", class)
				//fmt.Printf("marshalled %v", string(m))
				//log_err("err marrshalling", err)
			}
			_, err := session.ChannelMessageSendReply(channel.ID, ReplyText, message.Reference())
			log_err("Err Replying", err, false)

		}
		if args[1] == "homework" || args[1] == "coursework" || args[1] == "assignments" {
			var assignments []*classroom.CourseWork
			for _, class := range list_courses(srv, 10) {
				c, _ := srv.Courses.CourseWork.List(class.Id).Do()
				assignments = append(assignments, c.CourseWork...)

			}
			sort.Slice(assignments, func(i, j int) bool {
				i_time := time.Date(int(assignments[i].DueDate.Year), time.Month(assignments[i].DueDate.Month), int(assignments[i].DueDate.Day), int(assignments[i].DueTime.Hours), int(assignments[i].DueTime.Minutes), int(assignments[i].DueTime.Seconds), int(assignments[i].DueTime.Nanos), time.UTC)
				j_time := time.Date(int(assignments[j].DueDate.Year), time.Month(assignments[j].DueDate.Month), int(assignments[j].DueDate.Day), int(assignments[j].DueTime.Hours), int(assignments[j].DueTime.Minutes), int(assignments[j].DueTime.Seconds), int(assignments[j].DueTime.Nanos), time.UTC)
				return i_time.After(j_time)
			})
			for _, work := range assignments {
				fmt.Println(work.Title)

				//_, err := session.ChannelMessageSendReply(channel.ID, work.Title, message.Reference())
				//log_err("err replying", err, false)
			}

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
func load_token() string {
	// load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		log.Panicf("error loading .env file:%v", err)
		return ""
	}
	Token := os.Getenv("DISCORD_TOKEN")
	if Token == "" {
		log.Panicf("Token (or class id) not loaded. Make sure your .env file includes your Discord Token & Class_ID")
		return ""
	} else {
		return Token
	}
}
