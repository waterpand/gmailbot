package gmail

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var gmailClient *gmail.Service

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

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

func getService() *gmail.Service {
	if gmailClient == nil {
		b, err := ioutil.ReadFile("credentials.json")
		if err != nil {
			log.Fatalf("Unable to read client secret file: %v", err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
		if err != nil {
			log.Fatalf("Unable to parse client secret file to config: %v", err)
		}
		client := getClient(config)

		gmailClient, err = gmail.New(client)
		if err != nil {
			log.Fatalf("Unable to retrieve Gmail client: %v", err)
		}
	}

	return gmailClient
}

//GetNewestMessageID gets the Message.Id of ther first message in the list.
func GetNewestMessageID() string {
	srv := getService()
	user := "me"
	r, err := srv.Users.Messages.List(user).IncludeSpamTrash(false).MaxResults(1).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve Messages: %v", err)
	}
	if len(r.Messages) == 0 {
		fmt.Println("No Message found.")
		return ""
	}
	return r.Messages[0].Id
}

// GetMessage retrieves the detail of a message specified by ID.
func GetMessage(ID string) *gmail.Message {
	srv := getService()
	user := "me"
	m, err := srv.Users.Messages.Get(user, ID).Format("metadata").MetadataHeaders("Date", "Subject", "From").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve Message Id: %v", ID)
	}
	return m
}
