package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/mcuadros/go-defaults"
)

// AIRequest main API request for AI
type AIRequest struct {
	Type           string `default:"GENERATE" json:"type"`
	Style          string `default:"DEFAULT" json:"style"`
	Width          int    `default:"680" json:"width"`
	Height         int    `default:"1024" json:"height"`
	NumImages      int    `default:"1" json:"num_images"`
	NegativePrompt string `json:"negativePromptUnclip,omitempty"`

	GenerateParams struct {
		Query string `json:"query"`
	} `json:"generateParams"`
}

// AIModel model
type AIModel struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Version float64 `json:"version"`
	Type    string  `json:"type"`
}

// NewAIClient return new client for AI
func NewAIClient(cfg *Config) *AIClient {
	c := new(AIClient)
	defaults.SetDefaults(c)
	c.http = &http.Client{}
	c.cfg = cfg
	return c
}

// AIClient client for AI
type AIClient struct {
	Key    string
	Secret string

	ModelID int

	http *http.Client
	cfg  *Config
}

// AIRunResponse response run
type AIRunResponse struct {
	TaskID string `json:"uuid"`
	Status string `json:"status"`
}

// AIWaitResponse wait response
type AIWaitResponse struct {
	TaskID   string   `json:"uuid"`
	Status   string   `json:"status"`
	Images   [][]byte `json:"images"`
	Error    string   `json:"errorDescription"`
	Censored bool     `json:"censored"`
}

// GenImage generate one image
func (c *AIClient) genImage(profile *Profile) ([]byte, error) {

	if err := c.getModel(); err != nil {
		return nil, err
	}

	log.Printf("Модель получена %d (%d)", c.ModelID, profile.Telegram.UserID)

	task, err := c.runAI(profile)
	if err != nil {
		return nil, err
	}

	return c.waitTask(task, profile, c.cfg.AI.WaitTimeout)
}

// GenImages generate some images
func (c *AIClient) GenImages(profile *Profile) ([][]byte, error) {
	c.Key = profile.Access.Key
	c.Secret = profile.Access.Secret

	type iRes struct {
		Image []byte
		Error error
	}

	threadsPerClient := c.cfg.AI.ThreadsPerClient
	for _, a := range c.cfg.App.Admins {
		if a == profile.Telegram.UserID {
			threadsPerClient = c.cfg.AI.ThreadsPerAdmin
			log.Printf("admin detected, use more threads (%d)", threadsPerClient)
			break
		}
	}

	taskChan := make(chan bool, profile.Task.Count+16+threadsPerClient)
	resChan := make(chan *iRes, profile.Task.Count)

	for i := 0; i < profile.Task.Count; i++ {
		taskChan <- true

	}
	for i := 0; i < threadsPerClient; i++ {
		taskChan <- false
		go func() {
			for <-taskChan {
				log.Printf("Начата генерация одного изображения для пользователя %d",
					profile.Telegram.UserID)
				img, err := c.genImage(profile)
				resChan <- &iRes{img, err}
			}
		}()
	}

	images := make([][]byte, 0, profile.Task.Count)
	errors := make([]string, 0, profile.Task.Count)
	for i := 0; i < profile.Task.Count; i++ {
		res := <-resChan
		if res.Error != nil {
			errors = append(errors, res.Error.Error())
		} else {
			images = append(images, res.Image)
		}
		log.Printf("Получены резульаты ошибок: %d, изображений: %d",
			len(errors), len(images))
	}
	if len(images) > 0 {
		return images, nil
	}
	return nil, fmt.Errorf("ничего не получилось:\n\t%s",
		strings.Join(errors, "\n\t"))

}

func (c *AIClient) getModel() error {
	if c.ModelID != 0 {
		return nil
	}
	req, err := http.NewRequest("GET", "https://api-key.fusionbrain.ai/key/api/v1/models", nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-Key", fmt.Sprintf("Key %s", c.Key))
	req.Header.Add("X-Secret", fmt.Sprintf("Secret %s", c.Secret))

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case 401:
		return fmt.Errorf("wrong key or secret")
	case 200:
	default:
		return fmt.Errorf("Can't receive model")
	}

	models := []AIModel{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&models); err != nil {
		return err
	}

	if len(models) < 1 {
		return fmt.Errorf("No models found")
	}
	c.ModelID = models[0].ID
	return nil
}

func (c *AIClient) runAI(profile *Profile) (string, error) {

	aiReq := new(AIRequest)
	defaults.SetDefaults(aiReq)

	aiReq.Width = profile.Image.Width
	aiReq.Height = profile.Image.Height
	aiReq.NegativePrompt = profile.Task.Negative
	aiReq.GenerateParams.Query = profile.Task.Positive

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Type":        []string{"application/json"},
		"Content-Disposition": []string{`form-data; name="params"`},
	})
	if err != nil {
		return "", err
	}
	if paramsData, err := json.Marshal(aiReq); err == nil {
		part.Write(paramsData)
	} else {
		return "", err
	}
	writer.WriteField("model_id", fmt.Sprintf("%d", c.ModelID))
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api-key.fusionbrain.ai/key/api/v1/text2image/run", payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Key", fmt.Sprintf("Key %s", c.Key))
	req.Header.Add("X-Secret", fmt.Sprintf("Secret %s", c.Secret))
	req.Header.Add("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", writer.Boundary()))
	req.Header.Add("Content-Length", fmt.Sprintf("%d", len(payload.String())))

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	switch resp.StatusCode {
	case 401:
		return "", fmt.Errorf("wrong key or secret")
	case 200, 201:
	default:
		return "", fmt.Errorf("Can't run process: %d", resp.StatusCode)
	}

	runRes := AIRunResponse{}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&runRes); err != nil {
		return "", err
	}

	if runRes.Status != "INITIAL" {
		return "", fmt.Errorf("Non initial status for task: %v", runRes)
	}

	log.Printf("Задание составлено: %v (%d)", runRes, profile.Telegram.UserID)

	return runRes.TaskID, nil
}

func (c *AIClient) waitTask(task string, profile *Profile, timeout int) ([]byte, error) {

	started := time.Now()

	for attempt := 0; ; attempt++ {
		now := time.Now()

		if now.Sub(started) > time.Duration(timeout)*time.Second {
			break
		}

		time.Sleep(time.Second*1 + time.Second*time.Duration(rand.Intn(8)))
		// if attempt > 0 {
		// log.Printf("Продолжаем ожидать %d (%3.2f)",
		// profile.Telegram.UserID,
		// float64(attempt)*100/float64(attempts))
		// }
		var (
			decoder *json.Decoder
			ws      AIWaitResponse
		)

		url := fmt.Sprintf("https://api-key.fusionbrain.ai/key/api/v1/text2image/status/%s", task)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("Cant make http-request: %s", err)
		}
		req.Header.Add("X-Key", fmt.Sprintf("Key %s", c.Key))
		req.Header.Add("X-Secret", fmt.Sprintf("Secret %s", c.Secret))
		resp, err := c.http.Do(req)
		if err != nil {
			continue
		}
		switch resp.StatusCode {
		case 401:
			return nil, fmt.Errorf("wrong key or secret")
		case 200:
		default:
			log.Printf("Код ответа ожидания не 200: %d (%d)",
				resp.StatusCode, profile.Telegram.UserID)
			continue
		}

		ws = AIWaitResponse{}
		decoder = json.NewDecoder(resp.Body)
		if err := decoder.Decode(&ws); err != nil {
			continue
		}

		switch ws.Status {
		case "INITIAL":
			continue
		case "PROCESSING":
			continue

		case "FAIL":
			return nil, fmt.Errorf("Can't generate image: %s", ws.Error)
		case "DONE":
			if ws.Censored {
				return nil, fmt.Errorf("Цензура не пропустила")
			}
			return ws.Images[0], nil
		default:
			log.Printf("Неизвестный статус задачи %s (%d)", ws.Status, profile.Telegram.UserID)
		}
	}

	return nil, fmt.Errorf("Timeout exceeded")
}
