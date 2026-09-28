package source

import (
	"context"
	"os"

	"github.com/hat-y/Hather/internal/model"
)

// Local selects the caller-provided path directly; it never searches remotely.
type Local struct{}

func (Local) Fetch(ctx context.Context, wall model.Wallpaper) (model.Artifact, error) {
	if err := ctx.Err(); err != nil {
		return model.Artifact{}, Error{"cancelled", err.Error()}
	}
	if wall.ImageURL == "" {
		return model.Artifact{}, Error{"invalid_input", "a local image path is required"}
	}
	info, err := os.Stat(wall.ImageURL)
	if err != nil || info.IsDir() {
		return model.Artifact{}, Error{"invalid_input", "local image is unreadable"}
	}
	return artifact(wall.ID, wall.ImageURL, "")
}
