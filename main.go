package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	_ "embed"

	"github.com/go-telegram/bot"
	"github.com/unera/bot-cover/dialog"
)

func usage() {
	fmt.Printf("Usage: %s path/to/config.yaml\n", os.Args[0])
}

// Send any text message to the bot after the bot has been started
func main() {

	if len(os.Args) < 2 {
		usage()
		os.Exit(-1)
	}

	cfg := loadConfig(os.Args[1])

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	iniitImageSystem()
	defer closeImageSystem()

	b, err := bot.New(cfg.Telegram.Bot)
	if err != nil {
		panic(err)
	}

	texts := initTexts()

	opts := []dialog.Option{
		dialog.WithHandler(func(d *dialog.Dialog, profile any) {
			coverDialog(d, profile, texts, cfg)
		}),
		dialog.WithInactiveTimeout(900),

		dialog.WithProfileLoader(profileLoader, cfg.App.ProfileDir),
		dialog.WithProfileStorer(profileStorer, cfg.App.ProfileDir),
	}

	ctx = dialog.InstallRootDialog(ctx, b, opts...)

	b.Start(ctx)
}
