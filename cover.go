package main

import (
	_ "embed"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/unera/bot-cover/dialog"
	"gopkg.in/yaml.v3"
)

func colorType(c string) int {
	re3 := regexp.MustCompile("^#?[0-9a-fA-F]{3}$")
	re4 := regexp.MustCompile("^#?[0-9a-fA-F]{4}$")
	re6 := regexp.MustCompile("^#?[0-9a-fA-F]{6}$")
	re8 := regexp.MustCompile("^#?[0-9a-fA-F]{8}$")

	if re3.Match([]byte(c)) {
		return 3
	}
	if re4.Match([]byte(c)) {
		return 4
	}
	if re6.Match([]byte(c)) {
		return 6
	}
	if re8.Match([]byte(c)) {
		return 8
	}
	return -1
}

//go:embed colors.yaml
var predefinedColorsData []byte
var predefinedColors map[string]bool

func normalizeColor(value string) string {
	if predefinedColors == nil {
		predefinedColors = make(map[string]bool)
		yaml.Unmarshal(predefinedColorsData, predefinedColors)
	}

	if _, ok := predefinedColors[value]; ok {
		return value
	}

	if len(value) > 0 && value[0] != '#' {
		value = "#" + value
	}
	switch colorType(value) {
	case 6:
		value = value + "FF"
	case 8:
	case 3:
		value = string([]byte{
			value[0],
			value[1], value[1],
			value[2], value[2],
			value[3], value[3],
			'F', 'F',
		})
	case 4:
		value = string([]byte{
			value[0],
			value[1], value[1],
			value[2], value[2],
			value[3], value[3],
			value[4], value[4],
		})

	case -1:
		return ""
	}
	return value
}

func coverDialog(
	d *dialog.Dialog,
	profileRef any,
	texts *predefinedTexts,
	cfg *Config) {

	profile := profileRef.(*Profile)
	var il *ImageLabel

	text := d.GetText()
	reKey := regexp.MustCompile("^[0-9a-fA-F]{32}$")

	switch text {
	case "/start":
		d.SendHTML(texts.Make("first_start", profile))
		return
	case "/status":
		d.SendHTML(texts.Make("start", profile))
		return
	case "/width", "/height":
		d.SendHTML(texts.Make(text[1:], profile))
		value, err := strconv.ParseInt(d.GetText(), 10, 32)
		if err != nil {
			d.SendHTML(texts.Make("error", nil))
			return
		}

		if value < 100 || value > 1024 {
			d.SendHTML(texts.Make("wrong", "должно быть в диапазоне от 100 до 1024"))
			return
		}
		if text == "/width" {
			profile.Image.Width = int(value)
		} else {
			profile.Image.Height = int(value)
		}
		d.SendHTML(texts.Make("start", profile))
		return

	case "/top_text", "/bottom_text":
		if text == "/top_text" {
			il = &profile.Image.Top
		} else {
			il = &profile.Image.Bottom
		}
		d.SendHTML(texts.Make(text[1:], profile))
		switch value := d.GetText(); value {
		case "/clean":
			value = ""
			il.Text = ""
			d.SendHTML(texts.Make("start", profile))
			return
		case "/ok":
			value = il.Text
			if value == "" {
				d.SendHTML(texts.Make("start", profile))
				return
			}
		default:
			il.Text = value
		}
		d.SendHTML(texts.Make("start", profile))
		return

	case "/top_color", "/top_scolor", "/bottom_color", "/bottom_scolor":
		tdesc := new(struct {
			What  string
			Color *string
		})
		switch text {
		case "/top_color":
			tdesc.What = "цвет текста верхней надписи"
			tdesc.Color = &profile.Image.Top.Color
		case "/top_scolor":
			tdesc.What = "цвет границы текста верхней надписи"
			tdesc.Color = &profile.Image.Top.StrokeColor
		case "/bottom_color":
			tdesc.What = "цвет текста нижней надписи"
			tdesc.Color = &profile.Image.Bottom.Color
		case "/bottom_scolor":
			tdesc.What = "цвет границы текста нижней надписи"
			tdesc.Color = &profile.Image.Bottom.StrokeColor
		}
		d.SendHTML(texts.Make("color", tdesc))
		switch value := d.GetText(); value {
		case "/ok":
			return
		default:
			value = normalizeColor(value)
			if value == "" {
				d.SendHTML(texts.Make("color_error", nil))
				return
			}
			*tdesc.Color = value
		}
		d.SendHTML(texts.Make("start", profile))
		return

	case "/top_fontsize", "/bottom_fontsize":
		tdesc := new(struct {
			What string
			Size *int
		})
		if text == "/top_fontsize" {
			tdesc.What = "верхней надписи"
			tdesc.Size = &profile.Image.Top.Size
		} else {
			tdesc.What = "нижней надписи"
			tdesc.Size = &profile.Image.Bottom.Size
		}
		d.SendHTML(texts.Make("fontsize", tdesc))
		switch value := d.GetText(); value {
		case "/ok":
		default:
			if value, err := strconv.ParseInt(value, 10, 32); err == nil {
				if value < 3 || value > 33 {
					d.SendHTML(texts.Make("wrong", "должно быть в диапазоне от 3 до 33"))
					return
				}
				*tdesc.Size = int(value)
			}
		}
		d.SendHTML(texts.Make("start", profile))
		return

	case "/top", "/bottom":
		if text == "/top" {
			il = &profile.Image.Top
		} else {
			il = &profile.Image.Bottom
		}

		d.SendHTML(texts.Make("text_color", il))
		switch value := d.GetText(); value {
		case "/ok":
		default:
			value = normalizeColor(value)
			if value == "" {
				d.SendHTML(texts.Make("color_error", nil))
				return
			}
			il.Color = value
		}

		d.SendHTML(texts.Make("stroke_color", il))
		switch value := d.GetText(); value {
		case "/ok":
		default:
			value = normalizeColor(value)
			if value == "" {
				d.SendHTML(texts.Make("color_error", nil))
				return
			}
			il.StrokeColor = value
		}

		d.SendHTML(texts.Make("text_size", il))
		switch value := d.GetText(); value {
		case "/ok":
		default:
			if value, err := strconv.ParseInt(value, 10, 32); err == nil {
				if value < 3 || value > 33 {
					d.SendHTML(texts.Make("wrong", "должно быть в диапазоне от 3 до 33"))
					return
				}
				il.Size = int(value)
			}
		}

		d.SendHTML(texts.Make("start", profile))
		return
	case "/access":
		d.SendHTML(texts.Make("access", profile))
		return
	case "/access_keys":
		type accessTask struct {
			tpl   string
			value *string
		}
		for _, variant := range []accessTask{
			accessTask{"access_key", &profile.Access.Key},
			accessTask{"secret_key", &profile.Access.Secret},
		} {
			d.SendHTML(texts.Make(variant.tpl, profile))
			switch value := d.GetText(); value {
			case "/ok":
			case "/clean":
				*variant.value = ""
			default:
				if reKey.Match([]byte(value)) {
					*variant.value = value
				} else {
					d.SendHTML(texts.Make("wrong", "Строка должна быть длиной 32 символа."))
					return
				}
			}
		}
		d.SendHTML(texts.Make("start", profile))
		return
	case "/check":
		log.Printf("Preparing image")
		img := MakePredefinedImage(profile, cfg)

		log.Printf("Image prepared, size: %d bytes", len(img))
		d.SendAlbum(
			texts.Make("check"),
			&map[string][]byte{"example.png": img})
		return

	case "/ai_task", "/ai_avoid":
		d.SendHTML(texts.Make(text[1:], profile))
		switch value := d.GetText(); value {
		case "/ok":
		default:
			if len(value) > 1000 {
				d.SendHTML(texts.Make("too_long_text", len(value)))
				return
			}
			if text == "/ai_task" {
				profile.Task.Positive = value
			} else {
				profile.Task.Negative = value
			}
		}
		d.SendHTML(texts.Make("start", profile))
		return

	case "/top_font", "/bottom_font":
		tdesc := new(struct {
			What  string
			List  *map[string]string
			Value *string
		})
		tdesc.List = &cfg.App.Fonts
		if text == "/top_font" {
			tdesc.What = "верхней надписи"
			tdesc.Value = &profile.Image.Top.Font
		} else {
			tdesc.What = "Нижней надписи"
			tdesc.Value = &profile.Image.Bottom.Font
		}
		d.SendHTML(texts.Make("font", tdesc))

		switch value := d.GetText(); value {
		case "/ok":
		default:
			if len(value) > 0 {
				value = value[1:]
			}
			if _, ok := cfg.App.Fonts[value]; ok {
				*tdesc.Value = value
			} else {
				d.SendHTML(texts.Make("internal_error", "Неверный фонт"))
				return
			}
		}
		d.SendHTML(texts.Make("start", profile))
		return

	case "/run":
		if profile.Access.Key == "" || profile.Access.Secret == "" {
			d.SendHTML(texts.Make("access_error", profile))
			return
		}

		client := NewAIClient(cfg)
		d.SendHTML(texts.Make("please_wait", profile))
		imgList, err := client.GenImages(profile)
		if err != nil {
			d.SendHTML(texts.Make("internal_error", err))
			return
		}

		for len(imgList) > 0 {
			album := map[string][]byte{}
			for i := 0; i < 9 && len(imgList) > 0; i++ {
				name := fmt.Sprintf("image-%d.png", i)
				album[name] = MakeImage(imgList[0], profile, cfg)
				imgList = imgList[1:]
			}

			if len(imgList) > 0 {
				d.SendAlbum(texts.Make("part_done", profile), &album)
			} else {
				d.SendAlbum(texts.Make("done", profile), &album)
			}
		}

		return
	case "/faq":
		d.SendHTML(texts.Make("faq", profile))
		return

	}

	d.SendHTML(texts.Make("error", nil))
}
