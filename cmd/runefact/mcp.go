package main

import (
	mcpserver "github.com/vgalaktionov/runefact/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI agent integration",
	Long: `Start a Model Context Protocol server on stdio, exposing Runefact
operations as tools and resources for Claude Code, Claude Desktop,
and other MCP-capable AI agents.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}
		return mcpserver.StartMCPServer(root, cfg)
	},
}
