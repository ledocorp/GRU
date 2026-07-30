package ui

import (
	"errors"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// decodeThumbnailRaylib loads path off the main thread, downsamples to maxEdge, and
// returns a CPU image for main-thread GPU upload. Caller must rl.UnloadImage when done.
func decodeThumbnailRaylib(path string, maxEdge int) (*rl.Image, error) {
	if maxEdge < 1 {
		maxEdge = 256
	}

	rimg := rl.LoadImage(path)
	if rimg == nil || rimg.Width == 0 || rimg.Height == 0 {
		return nil, fmt.Errorf("raylib: could not load %q", path)
	}
	w, h := int(rimg.Width), int(rimg.Height)
	long := w
	if h > long {
		long = h
	}
	if long > maxEdge {
		var nw, nh int32
		if w >= h {
			nw = int32(maxEdge)
			nh = int32(float64(h) * float64(maxEdge) / float64(w))
		} else {
			nh = int32(maxEdge)
			nw = int32(float64(w) * float64(maxEdge) / float64(h))
		}
		if nw < 1 {
			nw = 1
		}
		if nh < 1 {
			nh = 1
		}
		rl.ImageResize(rimg, nw, nh)
		if rimg.Width == 0 || rimg.Height == 0 {
			rl.UnloadImage(rimg)
			return nil, errors.New("raylib: thumbnail resize failed")
		}
	}
	return rimg, nil
}

func thumbnailPNGBytes(path string, maxEdge int) ([]byte, error) {
	return exportThumbnailPNGRaylib(path, maxEdge)
}

func exportThumbnailPNGRaylib(path string, maxEdge int) ([]byte, error) {
	rimg, err := decodeThumbnailRaylib(path, maxEdge)
	if err != nil {
		return nil, err
	}
	defer rl.UnloadImage(rimg)
	data := rl.ExportImageToMemory(*rimg, ".png")
	if len(data) == 0 {
		return nil, errors.New("raylib: export png failed")
	}
	return data, nil
}
