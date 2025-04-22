package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
	"github.com/kwakuoseikwakye/go-mcps/internal/mcp"
)

type GithubServer struct {
	token  string
	client *github.Client
	ctx    context.Context
}

func (g *GithubServer) Name() string {
	return "github"
}

func (g *GithubServer) Connect(config map[string]string) error {
	token, ok := config["token"]
	if !ok || token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return errors.New("missing github token in config or environment")
	}
	g.token = token
	g.ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(g.ctx, ts)
	g.client = github.NewClient(tc)

	// Test auth with a simple call
	user, _, err := g.client.Users.Get(g.ctx, "")
	if err != nil {
		return fmt.Errorf("github authentication failed: %w", err)
	}
	fmt.Printf("Connected to GitHub as %s\n", user.GetLogin())
	return nil
}

func (g *GithubServer) ListContexts() ([]string, error) {
	repos, _, err := g.client.Repositories.List(g.ctx, "", nil)
	if err != nil {
		return nil, err
	}
	var contexts []string
	for _, repo := range repos {
		contexts = append(contexts, repo.GetFullName())
	}
	return contexts, nil
}

// ctx is repo name in "owner/repo" format, sends an issue comment to the latest issue
func (g *GithubServer) SendMessage(ctx, msg string) error {
	parts := strings.Split(ctx, "/")
	if len(parts) != 2 {
		return errors.New("context should be in owner/repo format")
	}
	owner, repo := parts[0], parts[1]
	issues, _, err := g.client.Issues.ListByRepo(g.ctx, owner, repo, nil)
	if err != nil || len(issues) == 0 {
		return fmt.Errorf("no issues found in %s: %w", ctx, err)
	}
	latest := issues[0]
	comment := &github.IssueComment{Body: &msg}
	_, _, err = g.client.Issues.CreateComment(g.ctx, owner, repo, latest.GetNumber(), comment)
	if err != nil {
		return fmt.Errorf("failed to send message as comment: %w", err)
	}
	fmt.Printf("Sent message to %s issue #%d: %s\n", ctx, latest.GetNumber(), msg)
	return nil
}

// Returns comments from the latest issue in the repo for demonstration
func (g *GithubServer) ReceiveMessage(ctx string) (<-chan mcp.Message, error) {
	parts := strings.Split(ctx, "/")
	if len(parts) != 2 {
		return nil, errors.New("context should be in owner/repo format")
	}
	owner, repo := parts[0], parts[1]
	ch := make(chan mcp.Message)
	go func() {
		defer close(ch)
		issues, _, err := g.client.Issues.ListByRepo(g.ctx, owner, repo, nil)
		if err != nil || len(issues) == 0 {
			return
		}
		latest := issues[0]
		comments, _, err := g.client.Issues.ListComments(g.ctx, owner, repo, latest.GetNumber(), nil)
		if err != nil {
			return
		}
		for _, comment := range comments {
			ch <- mcp.Message{
				Context: ctx,
				User:    comment.GetUser().GetLogin(),
				Text:    comment.GetBody(),
				Time:    comment.GetCreatedAt().String(),
			}
		}
	}()
	return ch, nil
}

func New() mcp.Server {
	return &GithubServer{}
}