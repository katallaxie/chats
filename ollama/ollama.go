package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/ollama/ollama/api"

	"github.com/katallaxie/pkg/conv"
	"github.com/katallaxie/prompts"
)

// Opts ...
type Opts struct {
	// BaseURL is the base URL.
	BaseURL string `json:"base_url"`
	// Timeout is the timeout.
	Timeout time.Duration `json:"timeout"`
	// Model is the model.
	Model string `json:"model"`
	// Client is the HTTP client.
	Client *http.Client `json:"-"`
	// Format is the format.
	Format json.RawMessage `json:"format"`
	// KeepAlive is the keep alive.
	KeepAlive bool `json:"keep_alive"`
	// Options is the options.
	Opts *api.Options `json:"options"`
}

// Opt ...
type Opt func(*Opts)

// Defaults ...
func Defaults() *Opts {
	return &Opts{}
}

var _ prompts.Prompter = (*Ollama)(nil)

// Ollama is a chat model.
type Ollama struct {
	client *api.Client
	opts   *Opts
}

// WithBaseURL configures the base URL.
func WithBaseURL(baseURL string) Opt {
	return func(o *Opts) {
		o.BaseURL = baseURL
	}
}

// WithModel configures the model.
func WithModel(model string) Opt {
	return func(o *Opts) {
		o.Model = model
	}
}

// New returns a new Ollama.
func New(opts ...Opt) (*Ollama, error) {
	options := Defaults()

	client := &http.Client{Timeout: options.Timeout}
	options.Client = client

	for _, opt := range opts {
		opt(options)
	}

	baseURL, err := url.Parse(options.BaseURL)
	if err != nil {
		return nil, err
	}

	model := new(Ollama)
	model.client = api.NewClient(baseURL, options.Client)
	model.opts = options

	return model, nil
}

// Complete is a completion.

// Generate ...
func (o *Ollama) Complete(ctx context.Context, prompt *prompts.Prompt) (*prompts.Completion, error) {
	complete := &prompts.Completion{
		Model:   prompt.Model,
		Choices: make([]prompts.CompletionChoice, 0),
	}

	req := &api.ChatRequest{
		Model: conv.String(prompt.Model),
	}

	for _, msg := range prompt.Messages {
		req.Messages = append(req.Messages, api.Message{
			Role:    conv.String(msg.GetRole()),
			Content: msg.GetContent(),
		})
	}

	fn := func(res api.ChatResponse) error {
		msg := prompts.CompletionChoice{
			Message: &prompts.GenericMessage{
				Role:    prompts.Role(res.Message.Role),
				Content: res.Message.Content,
			},
		}

		complete.Choices = append(complete.Choices, msg)

		return nil
	}

	err := o.client.Chat(ctx, req, fn)
	if err != nil {
		return nil, err
	}

	return complete, nil
}
