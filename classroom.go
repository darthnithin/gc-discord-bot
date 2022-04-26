package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/classroom/v1"
	"google.golang.org/api/option"
)

type discord_info struct {
	User    *discordgo.User
	Session *discordgo.Session
	msg_ref *discordgo.MessageReference
	channel *discordgo.Channel
	ids     chan string
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, info *discord_info) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := fmt.Sprintf("token_%v.json", info.User.ID)
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config, info)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config, info *discord_info) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	Auth_message := fmt.Sprintf("Go to the following link in your browser then reply with the provided authorization code: \n%v\n", authURL)
	info.Session.ChannelMessageSendReply(info.channel.ID, "Please Check your DM's for Authentication.", info.msg_ref)
	dmchannel, err := info.Session.UserChannelCreate(info.User.ID)
	if err != nil {
		log.Println(err)
	}
	dm, err := info.Session.ChannelMessageSend(dmchannel.ID, Auth_message)
	if err != nil {
		log.Println(err)
	}
	info.ids <- dm.ID
	var authCode string = <-tokenstream
	/* 	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	} */

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func Class(info *discord_info, result chan *classroom.Service) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read credentials file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	scopes := []string{classroom.ClassroomCoursesReadonlyScope, classroom.ClassroomAnnouncementsReadonlyScope, classroom.ClassroomCourseworkMeReadonlyScope}
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config, info)

	srv, err := classroom.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create classroom Client %v", err)
	}
	result <- srv

}
func announce(Class_id string, srv *classroom.Service) (*classroom.ListAnnouncementsResponse, error) {
	s, err := srv.Courses.Announcements.List(Class_id).PageSize(1).Do()
	if err == nil && len(s.Announcements) > 0 {
		return s, nil
	}
	return nil, err
}
func list_courses(srv *classroom.Service, PageSize int64) []*classroom.Course {
	r, err := srv.Courses.List().PageSize(PageSize).CourseStates("ACTIVE").Do()

	if err != nil {
		log.Fatalf("Unable to retrieve courses. %v", err)
	}
	return r.Courses
}
