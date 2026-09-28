package app

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/hat-y/Hather/internal/adapter"
	"github.com/hat-y/Hather/internal/config"
	"github.com/hat-y/Hather/internal/model"
)

type fakeSource struct {
	artifact model.Artifact
	err      error
}

func (f fakeSource) Fetch(context.Context, model.Wallpaper) (model.Artifact, error) {
	return f.artifact, f.err
}

type fakeEngine struct {
	palette model.Palette
	err     error
}

func (f fakeEngine) Extract(context.Context, model.Artifact) (model.Palette, error) {
	return f.palette, f.err
}

type fakeRunner struct {
	got    adapter.ApplyContext
	ids    []string
	result model.OperationResult
}

func (f *fakeRunner) Apply(_ context.Context, apply adapter.ApplyContext, ids []string) model.OperationResult {
	f.got, f.ids = apply, append([]string(nil), ids...)
	return f.result
}

func TestApplyGatesSharedPrerequisites(t *testing.T) {
	for _, service := range []Service{
		{Source: fakeSource{err: errors.New("unreadable")}, Engine: fakeEngine{}},
		{Source: fakeSource{}, Engine: fakeEngine{}},
		{Source: fakeSource{artifact: model.Artifact{Path: "image.png"}}, Engine: fakeEngine{err: errors.New("bad palette")}},
	} {
		runner := &fakeRunner{}
		service.Runner = runner
		got := service.Apply(context.Background(), ApplyRequest{Wallpaper: model.Wallpaper{ID: "local"}, Adapters: []string{"ghostty"}})
		if got.Status != model.OperationPrerequisiteFailure || len(runner.ids) != 0 {
			t.Fatalf("got %#v; runner called with %#v", got, runner.ids)
		}
	}
}

func TestPreviewAndApplyReportCancellation(t *testing.T) {
	service := Service{Source: fakeSource{err: context.Canceled}, Engine: fakeEngine{}, Runner: &fakeRunner{}}
	for _, got := range []model.OperationResult{
		service.Preview(context.Background(), model.Wallpaper{ID: "wall"}),
		service.Apply(context.Background(), ApplyRequest{Wallpaper: model.Wallpaper{ID: "wall"}}),
	} {
		if got.Status != model.OperationCancelled || len(got.Diagnostics) != 1 || got.Diagnostics[0] != context.Canceled.Error() {
			t.Fatalf("cancellation result = %#v", got)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got := Service{Source: fakeSource{err: errors.New("source observed cancellation")}, Engine: fakeEngine{}}.Preview(ctx, model.Wallpaper{ID: "wall"})
	if got.Status != model.OperationCancelled {
		t.Fatalf("cancelled context result = %#v", got)
	}
}

func TestApplyOrdersOnlyEnabledAdaptersAndPreservesResults(t *testing.T) {
	paths := config.PathsForHome(t.TempDir())
	runner := &fakeRunner{result: model.OperationResult{Status: model.OperationPartialFailure, AdapterResults: []model.AdapterResult{
		{AdapterID: "macos", Status: model.AdapterApplied}, {AdapterID: "vscode", Status: model.AdapterFailed},
	}}}
	service := Service{
		Source: fakeSource{artifact: model.Artifact{Path: "image.png"}},
		Engine: fakeEngine{palette: model.Palette{EngineVersion: "v1"}},
		Runner: runner,
		Store:  storeFor(paths),
	}
	got := service.Apply(context.Background(), ApplyRequest{Wallpaper: model.Wallpaper{ID: "local"}, Adapters: []string{"vscode", "macos"}})
	if want := []string{"vscode", "macos"}; !same(runner.ids, want) {
		t.Fatalf("runner request = %v, want %v", runner.ids, want)
	}
	if runner.got.Artifact.Path != "image.png" || runner.got.Palette.EngineVersion != "v1" {
		t.Fatalf("runner context = %#v", runner.got)
	}
	if got.Status != model.OperationPartialFailure || len(got.AdapterResults) != 2 || got.AdapterResults[0].Status != model.AdapterApplied {
		t.Fatalf("sibling results were not retained: %#v", got)
	}
	if _, err := os.Stat(paths.StateFile); err != nil {
		t.Fatalf("successful sibling work was not persisted: %v", err)
	}
}

func storeFor(paths config.Paths) *config.Store {
	store := config.NewStore(paths)
	return &store
}

func same(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
