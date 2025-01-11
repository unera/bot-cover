package dialog

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// Handler is a user defined dialog handler
type Handler func(d *Dialog, profile any)

type dialogListCacheKey string

const (
	cacheKey = dialogListCacheKey("root-dialog")
	mutexKey = dialogListCacheKey("root-mutex")
)

// Dialog structure
type Dialog struct {
	id              string
	chatID          int64
	userID          int64
	bot             *bot.Bot
	handler         Handler
	ch              chan *models.Update
	inactiveTimeout int

	profileLoader     ProfileLoader
	profileLoaderOpts []any

	profileStorer     ProfileStorer
	profileStorerOpts []any

	context context.Context

	lastSent time.Time
}

func (d *Dialog) rateLimitCheck() {
	now := time.Now()

	if now.Sub(d.lastSent) < 10*time.Millisecond {
		pause := 10*time.Millisecond - now.Sub(d.lastSent)
		time.Sleep(pause)
	}
	d.lastSent = now
}

// ID returns unique dialog identifier
func (d *Dialog) ID() string {
	return d.id
}

func dialogByUpdate(b *bot.Bot, update *models.Update, opts ...Option) *Dialog {
	if update.Message == nil {
		return nil
	}
	d := &Dialog{
		id: fmt.Sprintf("%p:[%d]:[%d]",
			b, update.Message.Chat.ID, update.Message.From.ID),
		inactiveTimeout: 900,
		bot:             b,
		chatID:          update.Message.Chat.ID,
		userID:          update.Message.From.ID,
		ch:              make(chan *models.Update, 128),
	}

	for _, o := range opts {
		o(d)
	}
	return d
}

func (d *Dialog) run(ctx context.Context) {
	log.Printf("Run new dialog goroutine: %s", d.id)

	cache := ctx.Value(cacheKey).(map[string]*Dialog)
	mutex := ctx.Value(mutexKey).(*sync.Mutex)

	mutex.Lock()
	cache[d.id] = d
	mutex.Unlock()

	var (
		profile any
		err     error
	)

	defer func() {
		if err := recover(); err != nil {
			log.Printf("recovered failed dialog %s: %s", d.id, err)
		}
		mutex.Lock()
		delete(cache, d.id)
		mutex.Unlock()
		if d.profileStorer != nil {
			err = d.profileStorer("bot", d.chatID, d.userID, profile, d.profileStorerOpts...)
			if err != nil {
				log.Printf("Error while write profile (%d:%d): %s",
					d.chatID, d.userID, err)
			}
		}

		log.Printf("Finished dialog goroutine: %s", d.id)
	}()

	if d.profileLoader != nil {
		profile, err = d.profileLoader("bot", d.chatID, d.userID, d.profileLoaderOpts...)
		if err != nil {
			panic(err)
		}
	}

	d.handler(d, profile)
}

// InstallRootDialog plugin the dialog system into telegram bot
func InstallRootDialog(ctx context.Context, b *bot.Bot, opts ...Option) context.Context {

	if ctx.Value(cacheKey) == nil {
		ctx = context.WithValue(ctx, cacheKey, make(map[string]*Dialog))
		ctx = context.WithValue(ctx, mutexKey, &sync.Mutex{})
	}

	h := func(ctx context.Context, b *bot.Bot, update *models.Update) {

		d := dialogByUpdate(b, update, opts...)
		if d == nil {
			return
		}

		cache := ctx.Value(cacheKey).(map[string]*Dialog)

		if dlg, ok := cache[d.id]; ok {
			d = dlg
		} else {
			go d.run(ctx)
		}
		d.context = ctx
		d.ch <- update
	}

	o := bot.WithDefaultHandler(h)
	o(b)

	return ctx
}
