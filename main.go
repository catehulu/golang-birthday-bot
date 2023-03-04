package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	BToken       string `json:"token"`
	DefaultMsg   string `json:"message,omitempty"`
	ChannelId    string `json:"channel_id"`
	BotChannelId string `json:"bot_channel_id,omitempty"`
}

type Birthday struct {
	Name    string `json:"name"`
	Message string `json:"message,omitempty"`
	DateStr string `json:"date,omitempty"`
	date    time.Time
}

var config Config
var birthdayList map[string][]Birthday

func init() {
	birthdayList = make(map[string][]Birthday)

	f, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Config file not found")
	}

	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)

	err = json.Unmarshal([]byte(byteValue), &config)
	if err != nil {
		log.Fatal("wrong json format for config")
	}

	if config.BToken == "" {
		log.Fatal("token not set!")
	}

	if config.ChannelId == "" {
		log.Fatal("channel id not set!")
	}

	f2, err := os.Open("birthday.json")
	if err != nil {
		log.Fatal("Config file not found")
	}
	defer f2.Close()

	temp := []Birthday{}
	byteValue2, _ := ioutil.ReadAll(f2)

	err = json.Unmarshal([]byte(byteValue2), &temp)
	if err != nil {
		log.Fatal("wrong json format for config")
	}

	for _, v := range temp {
		birthday := Birthday{
			Name:    v.Name,
			Message: v.Message,
			DateStr: v.DateStr,
		}

		date, err := time.Parse("01-02", birthday.DateStr)
		if err != nil {
			log.Fatal("one of the date format is invalid")
		}

		birthday.date = date
		birthdayList[birthday.DateStr] = append(birthdayList[birthday.DateStr], birthday)
	}
}

func main() {

	dbot, err := discordgo.New("Bot " + config.BToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dbot.AddHandler(botReady)
	dbot.AddHandler(messageCreate)

	dbot.Identify.Intents = discordgo.IntentGuildMessages

	err = dbot.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sign

	dbot.Close()
}

func botReady(s *discordgo.Session, r *discordgo.Ready) {
	dMessage := "Happy Birthday!"
	if config.DefaultMsg != "" {
		dMessage = config.DefaultMsg
	}

	for {
		now := time.Now().Format("01-02")
		if persons, ok := birthdayList[now]; ok {
			for _, v := range persons {
				message := dMessage
				if v.Message != "" {
					message = v.Message
				}
				s.ChannelMessageSend(config.ChannelId, message)
			}
		}
		time.Sleep(time.Hour * 24)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID && config.BotChannelId != m.ChannelID {
		return
	}

	if m.Content == "Allo" {
		s.ChannelMessageSend(config.BotChannelId, "Who is he talking to")
	}
}
