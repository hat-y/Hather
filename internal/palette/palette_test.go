package palette

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hat-y/Hather/internal/model"
)

type expectedPalette struct {
	EngineVersion string `json:"engine_version"`
	Roles         struct {
		Background string `json:"background"`
		Foreground string `json:"foreground"`
		Muted      string `json:"muted"`
		Accent     string `json:"accent"`
	} `json:"roles"`
	Colors []string `json:"ordered_colors_16"`
}

func TestExtractIsVersionPinnedAndFormatInvariant(t *testing.T) {
	bytes, err := os.ReadFile("testdata/expected-palette.json")
	if err != nil {
		t.Fatal(err)
	}
	var want expectedPalette
	if err := json.Unmarshal(bytes, &want); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"fixture.png", "fixture.jpg", "fixture.webp"} {
		artifact := fixture(t, filepath.Join("testdata", name))
		got, err := New().Extract(context.Background(), artifact)
		if err != nil {
			t.Fatalf("Extract(%s): %v", name, err)
		}
		again, _ := New().Extract(context.Background(), artifact)
		if !reflect.DeepEqual(got, again) {
			t.Fatalf("Extract(%s) was not deterministic", name)
		}
		if got.EngineVersion != want.EngineVersion || got.Background != want.Roles.Background || got.Foreground != want.Roles.Foreground || got.Muted != want.Roles.Muted || got.Accent != want.Roles.Accent || !reflect.DeepEqual(got.Colors[:], want.Colors) {
			t.Fatalf("Extract(%s) = %#v, want %#v", name, got, want)
		}
		if got.SourceArtifactSHA256 != artifact.SHA256 || len(got.Colors) != 16 {
			t.Fatalf("Extract(%s) lost source identity or color order", name)
		}
		if contrast(got.Foreground, got.Background) < 4.5 || contrast(got.Muted, got.Background) < 3 {
			t.Fatalf("Extract(%s) did not meet contrast targets", name)
		}
	}
}

func TestExtractReportsContrastLimitationForLowContrastFixture(t *testing.T) {
	path := filepath.Join(t.TempDir(), "low-contrast.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	image := image.NewRGBA(image.Rect(0, 0, 1, 1))
	image.SetRGBA(0, 0, color.RGBA{0x88, 0x88, 0x88, 0xff})
	if err := png.Encode(file, image); err != nil {
		t.Fatal(err)
	}
	file.Close()
	palette, err := New().Extract(context.Background(), fixture(t, path))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(palette.Diagnostics, " "), "contrast limitation") {
		t.Fatalf("diagnostics = %q, want contrast limitation", palette.Diagnostics)
	}
}

func TestExtractClassifiesUnsupportedFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fixture.heic")
	if err := os.WriteFile(path, []byte("not an image"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := New().Extract(context.Background(), fixture(t, path))
	if Class(err) != "unsupported_format" {
		t.Fatalf("Class(%v) = %q, want unsupported_format", err, Class(err))
	}
}

func fixture(t *testing.T, path string) model.Artifact {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(bytes)
	return model.Artifact{Path: path, SHA256: hex.EncodeToString(sum[:])}
}
