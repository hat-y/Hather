package model

import (
	"reflect"
	"testing"
)

func TestStableModelFields(t *testing.T) {
	wallpaper := Wallpaper{SourceKind: "local", ID: "id", Title: "title", PageURL: "page", ImageURL: "image", Attribution: "credit", Metadata: map[string]string{"path": "/tmp/a.png"}}
	artifact := Artifact{Path: "/tmp/a.png", SHA256: "sha", SourceID: wallpaper.ID, ContentType: "image/png", Size: 42}
	palette := Palette{EngineVersion: "v1", SourceArtifactSHA256: artifact.SHA256, Background: "#000000", Foreground: "#ffffff", Muted: "#999999", Accent: "#ff0000", Colors: [16]string{"#000000"}, Diagnostics: []string{"limited"}}
	result := AdapterResult{AdapterID: "ghostty", Status: AdapterApplied, ErrorClass: "", ChangedPaths: []string{"/tmp/config"}, GeneratedArtifacts: []string{"/tmp/theme"}, FollowUp: []string{"reload"}, Diagnostics: []string{"safe"}}
	operation := OperationResult{Status: OperationComplete, Wallpaper: wallpaper, Artifact: artifact, Palette: palette, RequestedAdapters: []string{"ghostty"}, AdapterResults: []AdapterResult{result}, Attribution: wallpaper.Attribution}

	if operation.Palette.EngineVersion != "v1" || len(operation.Palette.Colors) != 16 || operation.AdapterResults[0].Status != AdapterApplied {
		t.Fatalf("stable fields were not retained: %#v", operation)
	}
}

func TestStableStatuses(t *testing.T) {
	if got, want := []AdapterStatus{AdapterApplied, AdapterSkipped, AdapterUnavailable, AdapterConflict, AdapterFailed}, []AdapterStatus{"applied", "skipped", "unavailable", "conflict", "failed"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("adapter statuses = %q, want %q", got, want)
	}
	if got, want := []OperationStatus{OperationComplete, OperationPartialFailure, OperationPrerequisiteFailure, OperationCancelled}, []OperationStatus{"complete", "partial_failure", "prerequisite_failure", "cancelled"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("operation statuses = %q, want %q", got, want)
	}
}
