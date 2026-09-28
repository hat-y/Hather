package palette

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hat-y/Hather/internal/model"
	_ "golang.org/x/image/webp"
)

const version = "v1"

type Engine struct{}

type Error struct{ Class, Message string }

func (e Error) Error() string { return e.Message }

func Class(err error) string {
	if e, ok := err.(Error); ok {
		return e.Class
	}
	return "decode"
}

func New() Engine { return Engine{} }

func (Engine) Version() string { return version }

func (Engine) Extract(ctx context.Context, artifact model.Artifact) (model.Palette, error) {
	if err := ctx.Err(); err != nil {
		return model.Palette{}, Error{"cancelled", err.Error()}
	}
	if !supported(strings.ToLower(filepath.Ext(artifact.Path))) {
		return model.Palette{}, Error{"unsupported_format", "unsupported image format"}
	}
	file, err := os.Open(artifact.Path)
	if err != nil {
		return model.Palette{}, Error{"decode", "could not read image"}
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return model.Palette{}, Error{"decode", "could not decode image"}
	}
	colors, err := quantize(ctx, img)
	if err != nil {
		return model.Palette{}, err
	}

	palette := model.Palette{EngineVersion: version, SourceArtifactSHA256: artifact.SHA256}
	palette.Background = colors[0]
	palette.Foreground = contrasting(colors, palette.Background, 4.5, true)
	palette.Muted = contrasting(colors, palette.Background, 3, false)
	palette.Accent = colorful(colors, palette.Background)
	if palette.Foreground == palette.Background || palette.Muted == palette.Background {
		palette.Diagnostics = []string{"contrast limitation: source colors cannot meet readable text targets"}
	}
	for i := range palette.Colors {
		palette.Colors[i] = colors[i%len(colors)]
	}
	return palette, nil
}

func supported(extension string) bool {
	return extension == ".jpg" || extension == ".jpeg" || extension == ".png" || extension == ".webp"
}

type bucket struct{ r, g, b uint8 }

type counted struct {
	bucket
	count int
}

func quantize(ctx context.Context, img image.Image) ([]string, error) {
	bounds := img.Bounds()
	step := max(1, max(bounds.Dx(), bounds.Dy())/64)
	counts := map[bucket]int{}
	samples := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		if err := ctx.Err(); err != nil {
			return nil, Error{"cancelled", err.Error()}
		}
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			counts[bucket{uint8(r >> 15), uint8(g >> 15), uint8(b >> 15)}]++
			samples++
		}
	}
	bins := make([]counted, 0, len(counts))
	for key, count := range counts {
		if count >= max(1, samples/16) {
			bins = append(bins, counted{key, count})
		}
	}
	sort.Slice(bins, func(i, j int) bool {
		if bins[i].count != bins[j].count {
			return bins[i].count > bins[j].count
		}
		if bins[i].r != bins[j].r {
			return bins[i].r < bins[j].r
		}
		if bins[i].g != bins[j].g {
			return bins[i].g < bins[j].g
		}
		return bins[i].b < bins[j].b
	})
	if len(bins) == 0 {
		return nil, Error{"decode", "image has no usable colors"}
	}
	colors := make([]string, len(bins))
	for i, bin := range bins {
		colors[i] = fmt.Sprintf("#%02X%02X%02X", bin.r*192+32, bin.g*192+32, bin.b*192+32)
	}
	return colors, nil
}

func contrasting(colors []string, background string, target float64, highest bool) string {
	chosen, score := background, -1.0
	for _, color := range colors {
		value := contrast(color, background)
		if value >= target && ((highest && value > score) || (!highest && (score < 0 || value < score))) {
			chosen, score = color, value
		}
	}
	return chosen
}

func colorful(colors []string, background string) string {
	chosen, score := background, -1.0
	for _, color := range colors[1:] {
		r, g, b := rgb(color)
		value := float64(max(r, max(g, b)) - min(r, min(g, b)))
		if value > score {
			chosen, score = color, value
		}
	}
	return chosen
}

func contrast(left, right string) float64 {
	one, two := luminance(left), luminance(right)
	if one < two {
		one, two = two, one
	}
	return (one + 0.05) / (two + 0.05)
}

func luminance(hex string) float64 {
	r, g, b := rgb(hex)
	channel := func(value uint8) float64 {
		v := float64(value) / 255
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}

func rgb(hex string) (uint8, uint8, uint8) {
	var r, g, b uint8
	fmt.Sscanf(hex, "#%02X%02X%02X", &r, &g, &b)
	return r, g, b
}
