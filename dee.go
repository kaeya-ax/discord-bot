package main


import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
)

type Responses struct {
	Questions map[string]string `json:"questions"`
}

func main() {

	// Create a new Discord session using the provided bot token.
	newSession, err := discordgo.New("Bot " + "bot Token here")
	if err != nil {
		log.Println("Error creating Discord session,", err)
		return
	}

	// Open the file and defer its closing
	jsonFile, err := os.Open("./responses.json")
	if err != nil {
		log.Fatal(err)
	}

	defer jsonFile.Close()

	// Unmarshal the JSON into our Responses struct
	byteValue, _ := io.ReadAll(jsonFile)

	var responses Responses

	json.Unmarshal(byteValue, &responses)

	newSession.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		//checks if the message is a question and if it is, it sends the response
		for question, response := range responses.Questions {
			if m.Content == question {
				s.ChannelMessageSend(m.ChannelID, response)
				break
			}
		}

		for _, user := range m.Mentions {
			if user.ID == s.State.User.ID {
				// Remove the bot mention from the message content
				messageAfterMention := strings.Replace(m.Content, "<@!"+s.State.User.ID+">", " ", 1)

				// Check if the message after the mention is a question and if it is, send the response
				found := false
				for question, response := range responses.Questions {
					if strings.TrimSpace(messageAfterMention) == question {
						s.ChannelMessageSend(m.ChannelID, response)
						found = true
						break
					}
				}

				// If the question was not found, send a default response with a suggestion
				if !found {
					s.ChannelMessageSend(m.ChannelID, "I'm sorry, I didn't understand that.")
				}

				break
			}
		}

	})

	// In this example, we only care about receiving message events.
	newSession.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	// Open a websocket connection to Discord and begin listening.
	err = newSession.Open()
	if err != nil {
		log.Println("Error opening connection,", err)
		return
	}

	defer newSession.Close() // Close the connection when the function returns.

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running. Press CTRL-C to exit.")

	// Simple way to keep program running until CTRL-C is pressed.
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	<-sc
}
