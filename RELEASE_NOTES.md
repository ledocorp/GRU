# v0.9.2 — Linux + Android S1 on the public tip

Builds on **v0.9.0** (S1 ship default) and **v0.9.1** (`x11util` export).

## Highlights

- **Linux S1:** 1:1 present (client = window; `DisplayScale=1`); sticky Latin via outline → fx AA greyscale; chrome soft-bold Poppins Medium/SemiBold; Wayland/X11 chrome helpers; desktop packaging.
- **Android S1 (beta):** NativeActivity path — density/`Dp`, IME, SAF file pick, insets; GLES sticky text; skip desktop-only shaders/SSAA FBOs; `packaging/android/` + scripts.
- **Windows:** unchanged S1 story (DirectWrite ClearType, `segoeuib` chrome Bold). Verified after merges: hello/grudemo build, `go test ./ui/`, S1 smoke.
- Bold looking different on Linux vs Windows is **expected** (see cumulative notes / `docs/architecture/linux-s1-text.md`).

## Undo anchors (if needed)

| Ref | Points to |
|-----|-----------|
| `pre-platform-s1-merge` / `main-pre-platform-s1` | Before both merges (`v0.9.1`) |
| `pre-android-s1-merge` / `main-pre-android-s1` | After Linux only |

Safer than force-push: `git revert -m 1` on the merge commits (see maintainer undo doc).

## Module

```bash
go get github.com/ledocorp/gru@v0.9.2
go run ./cmd/hello
go run -tags grudemo .
```

Full 0.9.x rollup: [`RELEASE_NOTES_v0.9.md`](RELEASE_NOTES_v0.9.md).
