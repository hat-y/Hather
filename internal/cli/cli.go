package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/hat-y/Hather/internal/app"
	"github.com/hat-y/Hather/internal/config"
	"github.com/hat-y/Hather/internal/model"
)

const (
	ExitComplete     = 0
	ExitPartial      = 2
	ExitPrerequisite = 3
	ExitUsage        = 4
	ExitInternal     = 5
)

type Core interface {
	Search(context.Context, string) ([]model.Wallpaper, error)
	Preview(context.Context, model.Wallpaper) model.OperationResult
	Apply(context.Context, app.ApplyRequest) model.OperationResult
}

type CLI struct {
	Core                                     Core
	ConfigureWallhavenAPIKey                 func(string)
	Version                                  string
	DefaultQuery, QuerySource, AdapterSource string
	Out, Err                                 io.Writer
}

func (c CLI) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		return c.usage()
	}
	switch args[0] {
	case "search":
		return c.search(ctx, args[1:])
	case "preview":
		return c.preview(ctx, args[1:])
	case "apply":
		return c.apply(ctx, args[1:])
	case "help", "-h", "--help":
		return c.help()
	case "version", "--version":
		fmt.Fprintln(c.out(), c.Version)
		return ExitComplete
	default:
		return c.usage()
	}
}

func (c CLI) search(ctx context.Context, args []string) int {
	fs := c.flags("search")
	query, jsonOutput := fs.String("query", "", "Wallhaven query"), fs.Bool("json", false, "JSON output")
	apiKey := fs.String("wallhaven-api-key", "", "Wallhaven API key (not persisted)")
	if fs.Parse(args) != nil || c.Core == nil {
		return c.usage()
	}
	source := c.QuerySource
	if *query != "" {
		source = "flag"
	}
	if source == "" {
		source = "default"
	}
	resolved := config.Resolve(*query, "", c.DefaultQuery, "")
	if resolved == "" {
		return c.usage()
	}
	if c.ConfigureWallhavenAPIKey != nil {
		c.ConfigureWallhavenAPIKey(*apiKey)
	}
	walls, err := c.Core.Search(ctx, resolved)
	if err != nil {
		fmt.Fprintln(c.err(), err)
		return ExitInternal
	}
	if *jsonOutput {
		json.NewEncoder(c.out()).Encode(walls)
		return ExitComplete
	}
	fmt.Fprintf(c.out(), "query source: %s\n", source)
	for _, wall := range walls {
		fmt.Fprintf(c.out(), "%s\t%s\t%s\t%s\n", wall.ID, wall.Title, wall.Metadata["resolution"], wall.ImageURL)
		c.renderSource(wall)
	}
	return ExitComplete
}

func (c CLI) preview(ctx context.Context, args []string) int {
	return c.operation(ctx, "preview", args, false)
}
func (c CLI) apply(ctx context.Context, args []string) int {
	return c.operation(ctx, "apply", args, true)
}

func (c CLI) operation(ctx context.Context, name string, args []string, apply bool) int {
	fs := c.flags(name)
	image, adapters := fs.String("image", "", "local image path"), fs.String("adapters", "", "comma-separated adapters")
	wallhavenID := fs.String("wallhaven-id", "", "Wallhaven result ID")
	wallhavenImageURL := fs.String("wallhaven-image-url", "", "Wallhaven result image URL")
	wallhavenPageURL := fs.String("wallhaven-page-url", "", "Wallhaven result page URL")
	resolution := fs.String("wallhaven-resolution", "", "Wallhaven result resolution")
	category := fs.String("wallhaven-category", "", "Wallhaven result category")
	purity := fs.String("wallhaven-purity", "", "Wallhaven result purity")
	apiKey := fs.String("wallhaven-api-key", "", "Wallhaven API key (not persisted)")
	jsonOutput := fs.Bool("json", false, "JSON output")
	if fs.Parse(args) != nil || c.Core == nil {
		return c.usage()
	}
	if c.ConfigureWallhavenAPIKey != nil {
		c.ConfigureWallhavenAPIKey(*apiKey)
	}
	wall := model.Wallpaper{SourceKind: "local", ID: *image, ImageURL: *image, Title: *image}
	if *image == "" && *wallhavenID != "" && *wallhavenImageURL != "" {
		wall = model.Wallpaper{SourceKind: "wallhaven", ID: *wallhavenID, Title: "Wallhaven " + *wallhavenID, PageURL: *wallhavenPageURL, ImageURL: *wallhavenImageURL, Metadata: map[string]string{"resolution": *resolution, "category": *category, "purity": *purity}}
		if wall.PageURL == "" {
			wall.PageURL = "https://wallhaven.cc/w/" + url.PathEscape(wall.ID)
		}
		wall.Attribution = "Wallhaven: " + wall.PageURL
	}
	if wall.ID == "" {
		return c.usage()
	}
	var result model.OperationResult
	if apply {
		result = c.Core.Apply(ctx, app.ApplyRequest{Wallpaper: wall, Adapters: split(*adapters)})
	} else {
		result = c.Core.Preview(ctx, wall)
	}
	c.render(result, *jsonOutput)
	if !*jsonOutput && apply {
		origin := c.AdapterSource
		if *adapters != "" {
			origin = "flag"
		}
		if origin != "" {
			fmt.Fprintf(c.out(), "adapters source: %s\n", origin)
		}
	}
	return ExitCode(result)
}

func ExitCode(result model.OperationResult) int {
	switch result.Status {
	case model.OperationComplete:
		return ExitComplete
	case model.OperationPartialFailure:
		return ExitPartial
	case model.OperationPrerequisiteFailure:
		return ExitPrerequisite
	case model.OperationCancelled:
		return ExitInternal
	default:
		return ExitInternal
	}
}

func (c CLI) render(result model.OperationResult, jsonOutput bool) {
	if jsonOutput {
		json.NewEncoder(c.out()).Encode(result)
		return
	}
	fmt.Fprintf(c.out(), "status: %s\n", result.Status)
	if result.Wallpaper.Title != "" {
		fmt.Fprintf(c.out(), "wallpaper: %s\n", result.Wallpaper.Title)
	}
	c.renderSource(result.Wallpaper)
	if result.Artifact.Path != "" {
		fmt.Fprintf(c.out(), "artifact: %s\nsha256: %s\ncontent type: %s\nsize: %d\n", result.Artifact.Path, result.Artifact.SHA256, result.Artifact.ContentType, result.Artifact.Size)
	}
	if result.Palette.Accent != "" {
		fmt.Fprintf(c.out(), "palette: %s\nbackground: %s\nforeground: %s\nmuted: %s\naccent: %s\n", result.Palette.Accent, result.Palette.Background, result.Palette.Foreground, result.Palette.Muted, result.Palette.Accent)
	}
	for _, diagnostic := range result.Palette.Diagnostics {
		fmt.Fprintln(c.out(), diagnostic)
	}
	for _, adapter := range result.AdapterResults {
		fmt.Fprintf(c.out(), "%s: %s\n", adapter.AdapterID, adapter.Status)
		if adapter.ErrorClass != "" {
			fmt.Fprintf(c.out(), "error class: %s\n", adapter.ErrorClass)
		}
		for _, details := range []struct {
			label  string
			values []string
		}{{"changed paths", adapter.ChangedPaths}, {"generated artifacts", adapter.GeneratedArtifacts}, {"diagnostics", adapter.Diagnostics}, {"follow-up", adapter.FollowUp}} {
			if len(details.values) > 0 {
				fmt.Fprintf(c.out(), "%s: %s\n", details.label, strings.Join(details.values, "; "))
			}
		}
	}
	for _, diagnostic := range result.Diagnostics {
		fmt.Fprintln(c.out(), diagnostic)
	}
}
func (c CLI) renderSource(wall model.Wallpaper) {
	for _, field := range []struct{ label, value string }{
		{"source", wall.SourceKind}, {"page", wall.PageURL}, {"attribution", wall.Attribution},
		{"resolution", wall.Metadata["resolution"]}, {"category", wall.Metadata["category"]}, {"purity", wall.Metadata["purity"]},
	} {
		if field.value != "" {
			fmt.Fprintf(c.out(), "%s: %s\n", field.label, field.value)
		}
	}
}
func (c CLI) flags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(c.err())
	return fs
}
func (c CLI) help() int {
	fmt.Fprintln(c.out(), "Usage: hather <command> [options]\n\nCommands: tui, search, preview, apply, version")
	return ExitComplete
}
func (c CLI) usage() int {
	fmt.Fprintln(c.err(), "usage: hather tui|search|preview|apply|version")
	return ExitUsage
}
func (c CLI) out() io.Writer {
	if c.Out != nil {
		return c.Out
	}
	return io.Discard
}
func (c CLI) err() io.Writer {
	if c.Err != nil {
		return c.Err
	}
	return io.Discard
}
func split(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}
