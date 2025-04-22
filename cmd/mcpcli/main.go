package main

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/kwakuoseikwakye/go-mcps/internal/mcp"
	"github.com/kwakuoseikwakye/go-mcps/internal/slack"
	"github.com/kwakuoseikwakye/go-mcps/internal/github"
)

var (
	serverName string
	config     map[string]string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "mcpcli",
		Short: "MCP CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Register servers once
			mcp.RegisterServer(slack.New())
			mcp.RegisterServer(github.New())
		},
	}

	rootCmd.PersistentFlags().StringVarP(&serverName, "server", "s", "", "Name of the MCP server (slack|github)")

	rootCmd.AddCommand(
		connectCmd(),
		listCmd(),
		sendCmd(),
		recvCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func connectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect",
		Short: "Connect to an MCP server",
		Run: func(cmd *cobra.Command, args []string) {
			srv, ok := mcp.GetServer(serverName)
			if !ok {
				fmt.Println("Unknown server:", serverName)
				return
			}
			// For demo: use env vars for tokens
			err := srv.Connect(make(map[string]string))
			if err != nil {
				fmt.Println("Connect error:", err)
			} else {
				fmt.Println("Connected to", serverName)
			}
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List contexts on server",
		Run: func(cmd *cobra.Command, args []string) {
			srv, ok := mcp.GetServer(serverName)
			if !ok {
				fmt.Println("Unknown server:", serverName)
				return
			}
			_ = srv.Connect(make(map[string]string)) // Connect silently for demo
			ctxs, err := srv.ListContexts()
			if err != nil {
				fmt.Println("List error:", err)
			} else {
				fmt.Println("Contexts:")
				for _, c := range ctxs {
					fmt.Println(" -", c)
				}
			}
		},
	}
}

func sendCmd() *cobra.Command {
	var ctx, msg string
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a message",
		Run: func(cmd *cobra.Command, args []string) {
			srv, ok := mcp.GetServer(serverName)
			if !ok {
				fmt.Println("Unknown server:", serverName)
				return
			}
			_ = srv.Connect(make(map[string]string))
			if err := srv.SendMessage(ctx, msg); err != nil {
				fmt.Println("Send error:", err)
			}
		},
	}
	cmd.Flags().StringVarP(&ctx, "context", "c", "", "Context/channel/repo")
	cmd.Flags().StringVarP(&msg, "message", "m", "", "Message text")
	cmd.MarkFlagRequired("context")
	cmd.MarkFlagRequired("message")
	return cmd
}

func recvCmd() *cobra.Command {
	var ctx string
	cmd := &cobra.Command{
		Use:   "recv",
		Short: "Receive messages",
		Run: func(cmd *cobra.Command, args []string) {
			srv, ok := mcp.GetServer(serverName)
			if !ok {
				fmt.Println("Unknown server:", serverName)
				return
			}
			_ = srv.Connect(make(map[string]string))
			msgch, err := srv.ReceiveMessage(ctx)
			if err != nil {
				fmt.Println("Receive error:", err)
				return
			}
			for msg := range msgch {
				fmt.Printf("[%s] %s: %s\n", msg.Time, msg.User, msg.Text)
			}
		},
	}
	cmd.Flags().StringVarP(&ctx, "context", "c", "", "Context/channel/repo")
	cmd.MarkFlagRequired("context")
	return cmd
}