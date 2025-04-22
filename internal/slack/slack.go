package slack

import (
	"errors"
	"fmt"
	"github.com/kwakuoseikwakye/go-mcps/internal/mcp"
	"github.com/slack-go/slack"
	"os"
)

type SlackServer struct {
	token  string
	client *slack.Client
}

func (s *SlackServer) Name() string {
	return "slack"
}

func (s *SlackServer) Connect(config map[string]string) error {
	token, ok := config["token"]
	if !ok || token == "" {
		token = os.Getenv("SLACK_TOKEN")
	}
	if token == "" {
		return errors.New("missing slack token in config or environment")
	}
	s.token = token
	s.client = slack.New(token)
	// Test auth
	authResp, err := s.client.AuthTest()
	if err != nil {
		return fmt.Errorf("slack authentication failed: %w", err)
	}
	fmt.Printf("Connected to Slack as %s\n", authResp.User)
	return nil
}

func (s *SlackServer) ListContexts() ([]string, error) {
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	}
	var contexts []string
	channels, _, err := s.client.GetConversations(params)
	if err != nil {
		return nil, err
	}
	for _, ch := range channels {
		contexts = append(contexts, "#"+ch.Name)
	}
	return contexts, nil
}

func (s *SlackServer) SendMessage(ctx, msg string) error {
	channel := ctx
	if channel[0] == '#' {
		channel = channel[1:]
	}
	channelID, err := s.resolveChannelID(channel)
	if err != nil {
		return err
	}
	_, _, err = s.client.PostMessage(channelID, slack.MsgOptionText(msg, false))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	fmt.Printf("Sent message to %s: %s\n", ctx, msg)
	return nil
}

func (s *SlackServer) ReceiveMessage(ctx string) (<-chan mcp.Message, error) {
	// For demo: poll the channel history once.
	// For production: use Slack RTM or events API.
	channel := ctx
	if channel[0] == '#' {
		channel = channel[1:]
	}
	channelID, err := s.resolveChannelID(channel)
	if err != nil {
		return nil, err
	}

	ch := make(chan mcp.Message)
	go func() {
		historyParams := slack.GetConversationHistoryParameters{
			ChannelID: channelID,
			Limit:     5,
		}
		history, err := s.client.GetConversationHistory(&historyParams)
		if err != nil {
			close(ch)
			return
		}
		for _, msg := range history.Messages {
			ch <- mcp.Message{
				Context: ctx,
				User:    msg.User,
				Text:    msg.Text,
				Time:    msg.Timestamp,
			}
		}
		close(ch)
	}()
	return ch, nil
}

func (s *SlackServer) resolveChannelID(channel string) (string, error) {
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	}
	channels, _, err := s.client.GetConversations(params)
	if err != nil {
		return "", err
	}
	for _, ch := range channels {
		if ch.Name == channel {
			return ch.ID, nil
		}
	}
	return "", fmt.Errorf("channel %s not found", channel)
}

func New() mcp.Server {
	return &SlackServer{}
}
