package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"os"
	"strings"
)

type Configuration struct {
	MytoptagService string   `json:"mytoptag-service"`
	BotToken        string   `json:"bot-token"`
	Admins          []string `json:"admins"`
}

var config Configuration

const ImportEndpoint = "import/"
const SuggestionEndpoint = "suggestion/"
const CategoryEndpoint = "category/"

type ImportQueueStatus struct {
	QueueSize    int      `json:"queueSize"`
	ImportedSize int      `json:"importedSize"`
	FailedSize   int      `json:"failedSize"`
	Queue        []string `json:"queue"`
	Imported     []string `json:"imported"`
	Failed       []string `json:"failed"`
}

type CategoryTagsList struct {
	Data []struct {
		Tag       string `json:"tag"`
		Category  string `json:"category"`
		SortOrder int    `json:"sortOrder"`
	} `json:"data"`
}

type Profiles struct {
	Profiles []string `json:"profiles"`
}

func main() {
	config = readConfig()
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, config.BotToken)
		if update.Message.IsCommand() && userIsAdmin(update.Message.From.UserName) {
			var reply = processCommand(update)
			msg.Text = reply
		} else if len(update.Message.Text) > 0 {
			var reply = processText(update)
			msg.Text = reply
		}
		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

func readConfig() Configuration {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}

func userIsAdmin(username string) bool {
	for _, u := range config.Admins {
		if username == u {
			return true
		}
	}
	return false
}

func processCommand(msg tgbotapi.Update) string {
	switch msg.Message.Command() {
	case "import":
		if userIsAdmin(msg.Message.From.UserName) {
			return doImportCommand(msg.Message.Text)
		} else {
			return doCategoryCommand(msg.Message.Text)
		}
	default:
		return doCategoryCommand(msg.Message.Text)
	}
}

func processText(msg tgbotapi.Update) string {
	var categories []string
	var showCategories = false
	for idx, e := range strings.Split(msg.Message.Text, " ") {
		if idx == 0 && e == "showcategories" {
			showCategories = true
		} else if len(e) > 2 {
			categories = append(categories, e)
		}
		if idx > 9 {
			break
		}
	}
	if len(categories) == 0 {
		return "Invalid categories"
	}
	resp, err := http.Get(config.MytoptagService + SuggestionEndpoint + strings.Join(categories, ","))
	if err != nil {
		log.Panic(err)
	}
	tags := CategoryTagsList{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&tags)
	if jsonErr != nil {
		log.Panic(jsonErr)
	}
	return prettyPrintTagList(tags, showCategories)
}

func doCategoryCommand(msg string) string {
	resp, err := http.Get(config.MytoptagService + CategoryEndpoint + msg)
	if err != nil {
		log.Panic(err)
	}
	tags := CategoryTagsList{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&tags)
	if jsonErr != nil {
		log.Panic(jsonErr)
	}
	return prettyPrintTagList(tags, false)
}

func doImportCommand(msg string) string {
	const importCommand = "/import"
	arguments := strings.Split(strings.Replace(msg, importCommand, "", 1), " ")
	if len(arguments) > 0 || len(arguments[0]) == 1 {
		return getImportQueueStatus()
	} else {
		return addProfilesToImport(arguments)
	}
}

func addProfilesToImport(profiles []string) string {
	payload, err := json.Marshal(&Profiles{Profiles: profiles})
	if err != nil {
		log.Panic(err)
	}
	log.Println(payload)
	req, err := http.NewRequest("POST", config.MytoptagService+ImportEndpoint, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	status := ImportQueueStatus{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&status)
	if jsonErr != nil {
		log.Panic(jsonErr)
	}
	return prettyPrintImportStatus(status)
}

func getImportQueueStatus() string {
	resp, err := http.Get(config.MytoptagService + ImportEndpoint)
	if err != nil {
		log.Panic(err)
	}
	status := ImportQueueStatus{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&status)
	if jsonErr != nil {
		log.Panic(jsonErr)
	}
	return prettyPrintImportStatus(status)
}

func prettyPrintImportStatus(status ImportQueueStatus) string {
	return fmt.Sprintf(
		"Import queue size: %d profiles\n"+
			"Imported: %d profiles\n"+
			"Failed to import: %d profiles",
		status.QueueSize,
		status.ImportedSize,
		status.FailedSize)
}

func prettyPrintTagList(tags CategoryTagsList, withCategory bool) string {
	if len(tags.Data) == 0 {
		return "No tags found"
	}
	var result strings.Builder
	for _, e := range tags.Data {
		var line string
		if withCategory {
			line = e.Category + ": #" + e.Tag + "\n"
		} else {
			line = "#" + e.Tag + "\n"
		}
		result.WriteString(line)
	}
	return result.String()
}
