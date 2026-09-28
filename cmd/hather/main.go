package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hat-y/Hather/internal/adapter"
	"github.com/hat-y/Hather/internal/adapter/ghostty"
	"github.com/hat-y/Hather/internal/adapter/herdr"
	"github.com/hat-y/Hather/internal/adapter/macos"
	"github.com/hat-y/Hather/internal/adapter/neovim"
	"github.com/hat-y/Hather/internal/adapter/vscode"
	"github.com/hat-y/Hather/internal/app"
	"github.com/hat-y/Hather/internal/cli"
	"github.com/hat-y/Hather/internal/config"
	"github.com/hat-y/Hather/internal/model"
	"github.com/hat-y/Hather/internal/palette"
	"github.com/hat-y/Hather/internal/platform"
	"github.com/hat-y/Hather/internal/source"
	"github.com/hat-y/Hather/internal/tui"
)

var version = "dev"

func main() {
	home, _ := os.UserHomeDir()
	program, err := newProgram(home, os.Getenv("HATHER_ENABLED_ADAPTERS"), platform.Executor{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "configuration:", err)
		os.Exit(cli.ExitUsage)
	}
	program.Out, program.Err = os.Stdout, os.Stderr
	os.Exit(run(context.Background(), os.Args[1:], program, program.Core, startTUI, os.Stderr))
}

func newProgram(home, environment string, executor platform.CommandExecutor) (cli.CLI, error) {
	store := config.NewStore(config.PathsForHome(home))
	settings, err := config.Load(store.Paths.ConfigFile)
	if err != nil {
		return cli.CLI{}, err
	}
	remote := &source.Wallhaven{APIKey: config.ResolveWallhavenAPIKey("", os.Getenv("WALLHAVEN_API_KEY")), Store: store, RequestPolicy: source.NewRequestPolicy()}
	service := app.Service{Source: sources{remote: remote}, Engine: palette.New(), Runner: newRunner(home, executor), Store: &store}
	core := configuredCore{Service: service, environment: environment, configured: settings.EnabledAdapters}
	queryEnv := os.Getenv("HATHER_QUERY")
	adapterSource := config.ResolveSource("", environment, "")
	if adapterSource == "default" && len(settings.EnabledAdapters) > 0 {
		adapterSource = "config"
	}
	return cli.CLI{Core: core, ConfigureWallhavenAPIKey: configureWallhavenAPIKey(remote, os.Getenv("WALLHAVEN_API_KEY")), Version: version, DefaultQuery: config.Resolve("", queryEnv, settings.Query, ""), QuerySource: config.ResolveSource("", queryEnv, settings.Query), AdapterSource: adapterSource}, nil
}

type configuredCore struct {
	app.Service
	environment string
	configured  []string
}

func (c configuredCore) Apply(ctx context.Context, request app.ApplyRequest) model.OperationResult {
	request.Adapters = config.ResolveAdapters(request.Adapters, c.environment, c.configured, nil)
	return c.Service.Apply(ctx, request)
}

type tuiRunner func(tea.Model) error

func run(ctx context.Context, args []string, program cli.CLI, core tui.Core, runner tuiRunner, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "tui" {
		if runner == nil {
			fmt.Fprintln(stderr, "tui: runner is not configured")
			return cli.ExitInternal
		}
		ui := tui.New(core)
		ui.DefaultQuery = program.DefaultQuery
		ui.Input.SetValue(program.DefaultQuery)
		if err := runner(ui); err != nil {
			fmt.Fprintf(stderr, "tui: %v\n", err)
			return cli.ExitInternal
		}
		return cli.ExitComplete
	}
	return program.Run(ctx, args)
}

func startTUI(model tea.Model) error {
	_, err := tea.NewProgram(model).Run()
	return err
}

func configureWallhavenAPIKey(remote *source.Wallhaven, environment string) func(string) {
	return func(explicit string) { remote.APIKey = config.ResolveWallhavenAPIKey(explicit, environment) }
}

type sources struct{ remote *source.Wallhaven }

func (s sources) Search(ctx context.Context, query string) ([]model.Wallpaper, error) {
	return s.remote.Search(ctx, query)
}
func (s sources) Fetch(ctx context.Context, wall model.Wallpaper) (model.Artifact, error) {
	if wall.SourceKind == "local" {
		return (source.Local{}).Fetch(ctx, wall)
	}
	return s.remote.Fetch(ctx, wall)
}

func newRunner(home string, executor platform.CommandExecutor) adapter.Runner {
	mac := macos.Adapter{Executor: executor}
	term := ghostty.Adapter{ConfigDir: filepath.Join(home, ".config", "ghostty"), BoundaryRoot: home}
	session := herdr.Adapter{ConfigPath: filepath.Join(home, ".config", "herdr", "config.toml"), BoundaryRoot: home}
	editor := neovim.Adapter{ConfigRoot: neovim.DefaultConfigRoot(home), BoundaryRoot: home}
	code := vscode.Adapter{SettingsPath: vscode.DefaultSettingsPath(home), BoundaryRoot: home}
	return adapter.Runner{Adapters: []adapter.Adapter{
		adapter.FuncAdapter{Name: mac.ID(), Fn: func(ctx context.Context, apply adapter.ApplyContext) model.AdapterResult {
			return mac.Apply(ctx, apply.Artifact, apply.Palette)
		}},
		adapter.FuncAdapter{Name: term.ID(), Fn: func(ctx context.Context, apply adapter.ApplyContext) model.AdapterResult {
			return term.Apply(ctx, apply.Palette)
		}},
		adapter.FuncAdapter{Name: session.ID(), Fn: func(ctx context.Context, apply adapter.ApplyContext) model.AdapterResult {
			return session.Apply(ctx, apply.Palette)
		}},
		adapter.FuncAdapter{Name: editor.ID(), Fn: func(ctx context.Context, apply adapter.ApplyContext) model.AdapterResult {
			return editor.Apply(ctx, apply.Palette)
		}},
		adapter.FuncAdapter{Name: code.ID(), Fn: func(ctx context.Context, apply adapter.ApplyContext) model.AdapterResult {
			return code.Apply(ctx, apply.Palette)
		}},
	}}
}
