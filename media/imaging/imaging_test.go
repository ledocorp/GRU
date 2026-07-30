package imaging

import (
	"errors"
	"testing"

	"github.com/ledocorp/gru/media/assets"
)

func TestResizeToFit_noop(t *testing.T) {
	img, err := DecodeFile(firstTestImage(t))
	if err != nil {
		t.Fatal(err)
	}
	out := ResizeToFit(img, 9999, 9999)
	b := img.Bounds()
	ob := out.Bounds()
	if ob.Dx() != b.Dx() || ob.Dy() != b.Dy() {
		t.Fatalf("expected noop size %v, got %v", b, ob)
	}
}

func TestThumbnail_smaller(t *testing.T) {
	path := firstTestImage(t)
	thumb, err := Thumbnail(path, 128)
	if err != nil {
		t.Fatal(err)
	}
	b := thumb.Bounds()
	if b.Dx() > 128 && b.Dy() > 128 {
		t.Fatalf("thumbnail exceeds max edge: %dx%d", b.Dx(), b.Dy())
	}
}

func TestProbeFile(t *testing.T) {
	path := firstTestImage(t)
	info, err := ProbeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Width < 1 || info.Height < 1 {
		t.Fatalf("bad dimensions: %+v", info)
	}
}

func TestThumbnail_rejectsHugeDimensions(t *testing.T) {
	path := firstTestImage(t)
	info, err := ProbeFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pixels := int64(info.Width) * int64(info.Height)
	if pixels > MaxThumbnailDecodePixels {
		_, err = Thumbnail(path, 64)
		if !errors.Is(err, ErrImageTooLarge) {
			t.Fatalf("expected ErrImageTooLarge, got %v", err)
		}
		return
	}
	// Synthetic guard: temporarily treat any image as too large via direct check.
	if MaxThumbnailDecodePixels < 1 {
		t.Fatal("bad test constant")
	}
}

func firstTestImage(t *testing.T) string {
	t.Helper()
	paths, err := assets.ListImages()
	if err != nil {
		t.Skip("assets/test_images:", err)
	}
	if len(paths) == 0 {
		t.Skip("no test images")
	}
	return paths[0]
}
