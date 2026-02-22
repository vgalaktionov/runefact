package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
)

func (ctx *ServerContext) handleProjectStatus(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	status := map[string]any{
		"project_name": ctx.Config.Project.Name,
		"root":         ctx.ProjectRoot,
		"output":       ctx.Config.Project.Output,
		"package":      ctx.Config.Project.Package,
		"defaults": map[string]any{
			"sprite_size": ctx.Config.Defaults.SpriteSize,
			"sample_rate": ctx.Config.Defaults.SampleRate,
			"bit_depth":   ctx.Config.Defaults.BitDepth,
		},
		"preview": map[string]any{
			"window_width":  ctx.Config.Preview.WindowWidth,
			"window_height": ctx.Config.Preview.WindowHeight,
			"pixel_scale":   ctx.Config.Preview.PixelScale,
		},
	}

	// Check if build output exists.
	outputDir := filepath.Join(ctx.ProjectRoot, ctx.Config.Project.Output)
	if info, err := os.Stat(outputDir); err == nil && info.IsDir() {
		status["built"] = true
	} else {
		status["built"] = false
	}

	b, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return nil, err
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "runefact://project/status",
			MIMEType: "application/json",
			Text:     string(b),
		},
	}, nil
}

func (ctx *ServerContext) handleManifest(_ context.Context, _ mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	manifestPath := filepath.Join(ctx.ProjectRoot, ctx.Config.Project.Output, "manifest.go")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "runefact://manifest",
				MIMEType: "application/json",
				Text:     `{"error": "no manifest found, run build first"}`,
			},
		}, nil
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "runefact://manifest",
			MIMEType: "text/x-go",
			Text:     string(data),
		},
	}, nil
}
