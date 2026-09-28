package model

type Wallpaper struct {
	SourceKind, ID, Title, PageURL, ImageURL, Attribution string
	Metadata                                              map[string]string
}

type Artifact struct {
	Path, SHA256, SourceID, ContentType string
	Size                                int64
}

type Palette struct {
	EngineVersion, SourceArtifactSHA256   string
	Background, Foreground, Muted, Accent string
	Colors                                [16]string
	Diagnostics                           []string
}

type AdapterStatus string

const (
	AdapterApplied     AdapterStatus = "applied"
	AdapterSkipped     AdapterStatus = "skipped"
	AdapterUnavailable AdapterStatus = "unavailable"
	AdapterConflict    AdapterStatus = "conflict"
	AdapterFailed      AdapterStatus = "failed"
)

type AdapterResult struct {
	AdapterID                                  string
	Status                                     AdapterStatus
	ErrorClass                                 string
	ChangedPaths, GeneratedArtifacts, FollowUp []string
	Diagnostics                                []string
}

type OperationStatus string

const (
	OperationComplete            OperationStatus = "complete"
	OperationPartialFailure      OperationStatus = "partial_failure"
	OperationPrerequisiteFailure OperationStatus = "prerequisite_failure"
	OperationCancelled           OperationStatus = "cancelled"
)

type OperationResult struct {
	Status            OperationStatus
	Wallpaper         Wallpaper
	Artifact          Artifact
	Palette           Palette
	RequestedAdapters []string
	AdapterResults    []AdapterResult
	Attribution       string
	Diagnostics       []string
}
