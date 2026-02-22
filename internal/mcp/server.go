package mcp

import (
	"sync"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/vgalaktionov/runefact/internal/config"
)

// ServerContext holds shared state for MCP tool and resource handlers.
type ServerContext struct {
	Config      *config.ProjectConfig
	ProjectRoot string
	BuildMu     sync.Mutex
}

// StartMCPServer creates and starts the MCP server on stdio.
func StartMCPServer(projectRoot string, cfg *config.ProjectConfig) error {
	ctx := &ServerContext{
		Config:      cfg,
		ProjectRoot: projectRoot,
	}

	s := server.NewMCPServer(
		"runefact",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(false, false),
	)

	registerTools(s, ctx)
	registerResources(s, ctx)

	return server.ServeStdio(s)
}

func registerTools(s *server.MCPServer, ctx *ServerContext) {
	s.AddTool(mcp.Tool{
		Name:        "runefact_build",
		Description: "Compile rune asset files into game-ready artifacts (PNG, JSON, WAV)",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"scope": map[string]any{
					"type":        "string",
					"enum":        []string{"all", "sprites", "maps", "audio"},
					"description": "Asset type scope to build",
				},
				"files": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Specific files to build (empty = all)",
				},
			},
		},
	}, ctx.handleBuild)

	s.AddTool(mcp.Tool{
		Name:        "runefact_validate",
		Description: "Check rune files for errors without producing output artifacts",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"files": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Specific files to validate (empty = all)",
				},
			},
		},
	}, ctx.handleValidate)

	s.AddTool(mcp.Tool{
		Name:        "runefact_inspect_sprite",
		Description: "Get sprite sheet metadata: sprite names, dimensions, frame counts",
		InputSchema: mcp.ToolInputSchema{
			Type:     "object",
			Required: []string{"file"},
			Properties: map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Sprite file name (e.g., player.sprite)",
				},
			},
		},
	}, ctx.handleInspectSprite)

	s.AddTool(mcp.Tool{
		Name:        "runefact_inspect_map",
		Description: "Get map metadata: dimensions, layers, tile counts, entities",
		InputSchema: mcp.ToolInputSchema{
			Type:     "object",
			Required: []string{"file"},
			Properties: map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Map file name (e.g., world.map)",
				},
			},
		},
	}, ctx.handleInspectMap)

	s.AddTool(mcp.Tool{
		Name:        "runefact_inspect_audio",
		Description: "Get audio metadata: duration, voices, instruments",
		InputSchema: mcp.ToolInputSchema{
			Type:     "object",
			Required: []string{"file"},
			Properties: map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Audio file name (e.g., laser.sfx or bgm.track)",
				},
			},
		},
	}, ctx.handleInspectAudio)

	s.AddTool(mcp.Tool{
		Name:        "runefact_list_assets",
		Description: "List all rune asset files in the project, optionally filtered by type",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]any{
				"type": map[string]any{
					"type":        "string",
					"enum":        []string{"palette", "sprite", "map", "instrument", "sfx", "track"},
					"description": "Filter by asset type",
				},
			},
		},
	}, ctx.handleListAssets)

	s.AddTool(mcp.Tool{
		Name:        "runefact_palette_colors",
		Description: "Get resolved palette colors for a palette file",
		InputSchema: mcp.ToolInputSchema{
			Type:     "object",
			Required: []string{"file"},
			Properties: map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Palette file name (e.g., default.palette)",
				},
			},
		},
	}, ctx.handlePaletteColors)

	s.AddTool(mcp.Tool{
		Name:        "runefact_format_help",
		Description: "Get documentation for a rune asset format",
		InputSchema: mcp.ToolInputSchema{
			Type:     "object",
			Required: []string{"format"},
			Properties: map[string]any{
				"format": map[string]any{
					"type":        "string",
					"enum":        []string{"palette", "sprite", "map", "instrument", "sfx", "track"},
					"description": "Format to get help for",
				},
			},
		},
	}, ctx.handleFormatHelp)
}

func registerResources(s *server.MCPServer, ctx *ServerContext) {
	s.AddResource(mcp.Resource{
		URI:         "runefact://project/status",
		Name:        "Project Status",
		Description: "Current project configuration and build status",
		MIMEType:    "application/json",
	}, ctx.handleProjectStatus)

	s.AddResource(mcp.Resource{
		URI:         "runefact://manifest",
		Name:        "Build Manifest",
		Description: "Current manifest data as JSON",
		MIMEType:    "application/json",
	}, ctx.handleManifest)
}
