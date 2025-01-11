package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mcuadros/go-defaults"
	"gopkg.in/yaml.v3"
)

// ImageLabel for image
type ImageLabel struct {
	Text        string `yaml:"text" default:"какой-то текст"`
	Color       string `yaml:"color" default:"white"`
	StrokeColor string `yaml:"stroke_color" default:"black"`
	Size        int    `yaml:"size" default:"15"`
	Font        string `yaml:"font" default:"dejavu"`
}

// Profile for user
type Profile struct {
	Telegram struct {
		BotID  string `yaml:"bot_id"`
		UserID int64  `yaml:"id"`
		ChatID int64  `yaml:"chat_id"`
	} `yaml:"telegram"`

	Task struct {
		Positive string `yaml:"positive" default:"Красивый вид из окна на море"`
		Negative string `yaml:"negative" default:"Ядовитые цвета"`
		Count    int    `yaml:"count" default:"18"`
	} `yaml:"task"`

	Access struct {
		Key    string `yaml:"key"`
		Secret string `yaml:"secret"`
	} `yaml:"access"`

	Image struct {
		Top    ImageLabel `yaml:"top"`
		Bottom ImageLabel `yaml:"bottom"`
		Width  int        `yaml:"width" default:"680"`
		Height int        `yaml:"height" default:"1024"`
	}
}

// Bytes serialize profile
func (p *Profile) Bytes() []byte {
	r, e := yaml.Marshal(*p)
	if e != nil {
		panic(e)
	}
	return r
}

func (p *Profile) String() string {
	return string(p.Bytes())
}

// BaseName returns name for database
func (p *Profile) BaseName() string {
	return fmt.Sprintf("%s-%d-%d.yaml", p.Telegram.BotID, p.Telegram.ChatID, p.Telegram.UserID)
}

func profileLoader(botID string, chatID int64, userID int64, opts ...any) (any, error) {

	profile := new(Profile)

	defaults.SetDefaults(profile)

	profile.Telegram.BotID = botID
	profile.Telegram.ChatID = chatID
	profile.Telegram.UserID = userID

	fileName := filepath.Join(opts[0].(string), profile.BaseName())
	if profileRaw, err := os.ReadFile(fileName); err == nil {
		if err := yaml.Unmarshal(profileRaw, profile); err != nil {
			log.Printf("Wrong file format %s: %s", fileName, err)
		}
	}
	profile.Telegram.BotID = botID
	profile.Telegram.ChatID = chatID
	profile.Telegram.UserID = userID

	return profile, nil
}

func profileStorer(botID string, chatID int64, userID int64, profileA any, opts ...any) error {

	profile := profileA.(*Profile)
	fileName := filepath.Join(opts[0].(string), profile.BaseName())
	progressName := fmt.Sprintf("%s.inprogress", fileName)

	err := os.WriteFile(progressName, profile.Bytes(), 0644)
	if err == nil {
		err = os.Rename(progressName, fileName)
		if err != nil {
			log.Printf("Error write %s: %s", fileName, err)
		}
	} else {
		log.Printf("Error write %s: %s", fileName, err)
	}

	return nil
}
