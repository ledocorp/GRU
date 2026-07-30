// Package imaging provides decode and thumbnail helpers for T2a (pure Go).
//
// Decode and resize run off the main thread via ui.SubmitAsyncBg; upload textures
// on the main thread only. Thumbnail probes dimensions first and refuses full
// decode when width×height exceeds MaxThumbnailDecodePixels to avoid OOM on large
// camera JPEGs in gallery demos.
package imaging

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"image/png"
	"os"

	_ "golang.org/x/image/webp"

	"golang.org/x/image/draw"
)

// MaxThumbnailDecodePixels is the largest width×height the pure-Go Thumbnail path
// will fully decode. Larger files return ErrImageTooLarge so callers can fall back
// to a GPU-library resize path (ui.Image.LoadThumbnailAsync uses raylib/stb).
const MaxThumbnailDecodePixels = 32 << 20 // 32 megapixels

// maxConcurrentThumbnailDecodes limits parallel full decodes (gallery demos).
const maxConcurrentThumbnailDecodes = 2

var thumbDecodeSem = make(chan struct{}, maxConcurrentThumbnailDecodes)

// AcquireThumbnailDecode blocks until a gallery decode slot is free (max 2 concurrent).
func AcquireThumbnailDecode() {
	thumbDecodeSem <- struct{}{}
}

// ReleaseThumbnailDecode releases a gallery decode slot.
func ReleaseThumbnailDecode() {
	<-thumbDecodeSem
}

// ErrImageTooLarge is returned by Thumbnail when ProbeFile reports dimensions
// above MaxThumbnailDecodePixels.
var ErrImageTooLarge = errors.New("imaging: image exceeds max decode size for thumbnail")

// Info describes a decoded file without loading full pixels into a texture.
type Info struct {
	Path       string
	Format     string
	Width      int
	Height     int
	FileBytes  int64
}

// ProbeFile opens path and reads image config (dimensions, format).
func ProbeFile(path string) (Info, error) {
	st, err := os.Stat(path)
	if err != nil {
		return Info{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return Info{}, err
	}
	defer f.Close()
	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return Info{}, err
	}
	return Info{
		Path:      path,
		Format:    format,
		Width:     cfg.Width,
		Height:    cfg.Height,
		FileBytes: st.Size(),
	}, nil
}

// DecodeFile loads the full image at path.
func DecodeFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

// Thumbnail decodes path and downsamples so the longest edge is at most maxEdge.
//
// Returns ErrImageTooLarge when the file dimensions exceed MaxThumbnailDecodePixels
// without decoding full pixels — use a smaller source or a dedicated streaming decoder.
func Thumbnail(path string, maxEdge int) (image.Image, error) {
	if maxEdge < 1 {
		maxEdge = 256
	}
	info, err := ProbeFile(path)
	if err != nil {
		return nil, err
	}
	if int64(info.Width)*int64(info.Height) > MaxThumbnailDecodePixels {
		return nil, fmt.Errorf("%w (%dx%d)", ErrImageTooLarge, info.Width, info.Height)
	}
	AcquireThumbnailDecode()
	defer ReleaseThumbnailDecode()
	img, err := DecodeFile(path)
	if err != nil {
		return nil, err
	}
	return ResizeToFit(img, maxEdge, maxEdge), nil
}

// ThumbnailPNG decodes path, downsamples to maxEdge, and returns PNG bytes for GPU upload.
func ThumbnailPNG(path string, maxEdge int) ([]byte, error) {
	thumb, err := Thumbnail(path, maxEdge)
	if err != nil {
		return nil, err
	}
	return EncodePNG(thumb)
}

// ResizeToFit returns a copy scaled down to fit within maxW×maxH (aspect preserved).
// Images already within bounds are returned unchanged.
func ResizeToFit(src image.Image, maxW, maxH int) image.Image {
	if maxW < 1 {
		maxW = 1
	}
	if maxH < 1 {
		maxH = 1
	}
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxW && h <= maxH {
		return src
	}
	scale := min(float64(maxW)/float64(w), float64(maxH)/float64(h))
	nw := int(float64(w) * scale)
	nh := int(float64(h) * scale)
	if nw < 1 {
		nw = 1
	}
	if nh < 1 {
		nh = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, b, draw.Over, nil)
	return dst
}

// EncodePNG encodes img as PNG bytes (for raylib LoadImageFromMemory on the main thread).
func EncodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
