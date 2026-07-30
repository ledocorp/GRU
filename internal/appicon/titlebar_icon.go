package appicon



import (

	"sync"



	rl "github.com/gen2brain/raylib-go/raylib"

)



var (

	titleBarIconMu    sync.Mutex

	titleBarIconTex   rl.Texture2D

	titleBarIconDark  bool

)



// InvalidateTitleBarIcon drops the cached title-bar texture (call after SetPreferDarkIcon).

func InvalidateTitleBarIcon() {

	titleBarIconMu.Lock()

	defer titleBarIconMu.Unlock()

	if titleBarIconTex.ID != 0 {

		rl.UnloadTexture(titleBarIconTex)

		titleBarIconTex = rl.Texture2D{}

	}

}



// DrawTitleBarIcon paints the four-square GRU icon (32px source, scaled to rect).

// Uses the dark-chrome variant when [SetPreferDarkIcon] is true.

func DrawTitleBarIcon(rect rl.Rectangle) {

	tex := titleBarIconTexture()

	if tex.ID == 0 {

		return

	}

	src := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))

	rl.DrawTexturePro(tex, src, rect, rl.Vector2{}, 0, rl.White)

}



func titleBarIconTexture() rl.Texture2D {

	dark := PreferDarkIcon()

	titleBarIconMu.Lock()

	defer titleBarIconMu.Unlock()

	if titleBarIconTex.ID != 0 && titleBarIconDark == dark {

		return titleBarIconTex

	}

	if titleBarIconTex.ID != 0 {

		rl.UnloadTexture(titleBarIconTex)

		titleBarIconTex = rl.Texture2D{}

	}

	img := loadPNG(activeWindow32PNG())

	if img == nil {

		return rl.Texture2D{}

	}

	defer rl.UnloadImage(img)

	titleBarIconTex = rl.LoadTextureFromImage(img)

	titleBarIconDark = dark

	return titleBarIconTex

}


