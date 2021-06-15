package main

import (
	"encoding/json"
	"fmt"
	"greeting/api"
	"greeting/gomail"
	"greeting/static"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	env "github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

// User for receive email
type User struct {
	Email string `json:"email"`
	Local string `json:"local"`
}

func main() {
	loadConfig()

	location, _ := time.LoadLocation("Asia/Shanghai")
	cJob := cron.New(cron.WithLocation(location))

	cronConfig := os.Getenv("MAIL_CRON")
	if cronConfig == "" {
		bashSendMail()
	} else {
		cJob.AddFunc(cronConfig, func() {
			bashSendMail()
		})
		cJob.Start()
		select {}
	}
}

func sendMail(content string, to string) {
	gomail.Config.Username = os.Getenv("MAIL_USERNAME")
	gomail.Config.Password = os.Getenv("MAIL_PASSWORD")
	gomail.Config.Host = os.Getenv("MAIL_HOST")
	gomail.Config.Port = os.Getenv("MAIL_PORT")
	gomail.Config.From = os.Getenv("MAIL_FROM")

	email := gomail.GoMail{
		To:      []string{to},
		Subject: os.Getenv("MAIL_SUBJECT"),
		Content: content,
	}
	if err := email.Send(); err != nil {
		log.Fatalf("email send error: %s", err)
	}
}

func getParts() map[string]interface{} {
	wrapMap := map[string]func() interface{}{
		"one":       func() interface{} { return api.GetOne() },
		"english":   func() interface{} { return api.GetEnglish() },
		"poem":      func() interface{} { return api.GetPoem() },
		"wallpaper": func() interface{} { return api.GetWallpaper() },
		//"trivia":    func() interface{} { return api.GetTrivia() },
	}

	wg := sync.WaitGroup{}
	parts := make(map[string]interface{})
	for name, getPart := range wrapMap {
		wg.Add(1)
		go func(key string, fn func() interface{}) {
			defer wg.Done()
			parts[key] = fn()
		}(name, getPart)
		wg.Wait()
	}
	return parts
}

func bashSendMail() {
	users := getUsers()
	if len(users) == 0 {
		return
	}
	parts := getParts()
	if isDev() {
		parts["weather"] = api.GetWeather(users[0].Local)
		html := generateHTML(static.HTML, parts)
		fmt.Println(html)
		return
	}
	//
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	for _, user := range users {
		wg.Add(1)
		go func(user User) {
			defer wg.Done()
			parts["weather"] = api.GetWeather(user.Local)
			lock.Lock()
			html := generateHTML(static.HTML, parts)
			lock.Unlock()
			sendMail(html, user.Email)
		}(user)
	}
	wg.Wait()
}

func generateHTML(html string, datas map[string]interface{}) string {
	for key, data := range datas {
		rDataKey := reflect.TypeOf(data)
		rDataVal := reflect.ValueOf(data)
		fieldNum := rDataKey.NumField()
		for i := 0; i < fieldNum; i++ {
			fName := rDataKey.Field(i).Name
			rValue := rDataVal.Field(i)

			var fValue string
			switch rValue.Interface().(type) {
			case string:
				fValue = rValue.String()
			case []string:
				fValue = strings.Join(rValue.Interface().([]string), "<br>")
			}

			mark := fmt.Sprintf("{{%s.%s}}", key, fName)
			html = strings.ReplaceAll(html, mark, fValue)
		}
	}
	return html
}

func getUsers() []User {
	var users []User
	userJSON := os.Getenv("MAIL_TO")
	if err := json.Unmarshal([]byte(userJSON), &users); err != nil {
		log.Fatalf("get users error: %s", err)
	}
	return users
}

func isDev() bool {
	return os.Getenv("MAIL_MODE") == "dev"
}

func loadConfig() {
	if err := env.Load(); err != nil {
		log.Fatalf("Load .env file error: %s", err)
	}
}
