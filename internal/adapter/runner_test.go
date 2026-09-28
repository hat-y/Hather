package adapter

import (
	"context"
	"reflect"
	"testing"

	"github.com/hat-y/Hather/internal/model"
)

type fakeAdapter struct {
	id     string
	result model.AdapterResult
	calls  *[]string
	got    *ApplyContext
}

func (a fakeAdapter) ID() string { return a.id }
func (a fakeAdapter) Apply(_ context.Context, apply ApplyContext) model.AdapterResult {
	*a.calls = append(*a.calls, a.id)
	*a.got = apply
	return a.result
}

func TestRunnerOrdersResultsAndRetainsSiblingArtifacts(t *testing.T) {
	var calls []string
	var got ApplyContext
	apply := ApplyContext{Artifact: model.Artifact{Path: "wallpaper.png"}, Palette: model.Palette{Accent: "#123456"}}
	runner := Runner{Adapters: []Adapter{
		fakeAdapter{id: "vscode", calls: &calls, got: &got, result: model.AdapterResult{Status: model.AdapterApplied}},
		fakeAdapter{id: "ghostty", calls: &calls, got: &got, result: model.AdapterResult{Status: model.AdapterFailed, ErrorClass: "write"}},
		fakeAdapter{id: "macos", calls: &calls, got: &got, result: model.AdapterResult{Status: model.AdapterApplied, GeneratedArtifacts: []string{"desktop"}}},
		fakeAdapter{id: "herdr", calls: &calls, got: &got, result: model.AdapterResult{Status: model.AdapterApplied}},
		fakeAdapter{id: "neovim", calls: &calls, got: &got, result: model.AdapterResult{Status: model.AdapterApplied}},
	}}

	result := runner.Apply(context.Background(), apply, []string{"vscode", "ghostty", "macos", "herdr", "neovim"})
	want := []string{"macos", "ghostty", "herdr", "neovim", "vscode"}
	if !reflect.DeepEqual(calls, want) || result.Status != model.OperationPartialFailure || !reflect.DeepEqual(result.AdapterResults[0].GeneratedArtifacts, []string{"desktop"}) || !reflect.DeepEqual(got, apply) {
		t.Fatalf("calls=%v result=%#v context=%#v", calls, result, got)
	}
}

func TestRunnerLeavesDisabledUntouchedAndReportsUnwiredAdapters(t *testing.T) {
	var calls []string
	var got ApplyContext
	runner := Runner{Adapters: []Adapter{fakeAdapter{id: "ghostty", calls: &calls, got: &got, result: model.AdapterResult{Status: model.AdapterApplied}}}}

	result := runner.Apply(context.Background(), ApplyContext{}, []string{"ghostty", "future"})
	if !reflect.DeepEqual(calls, []string{"ghostty"}) || len(result.AdapterResults) != 2 || result.AdapterResults[1].AdapterID != "future" || result.AdapterResults[1].Status != model.AdapterUnavailable || result.Status != model.OperationPartialFailure {
		t.Fatalf("calls=%v result=%#v", calls, result)
	}
}
