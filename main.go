package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	windowWidth  = 800
	windowHeight = 600
	windowTitle  = "SDL2 in Go"
	spriteHeight = 128
	spriteWidth  = 128
)

func initSDL() error {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return fmt.Errorf("Error initializing SDL2: %v", err)
	}
	err = img.Init(img.INIT_PNG)
	if err != nil {
		return fmt.Errorf("Error initializing SDL_image: %v", err)
	}
	err = ttf.Init()
	if err != nil {
		return fmt.Errorf("Error initializing SDL_ttf: %v", err)
	}
	err = mix.Init(mix.INIT_OGG)
	if err != nil {
		return fmt.Errorf("Error initializing SDL_mixer: %v", err)
	}
	return nil
}

func closeSDL() {
	ttf.Quit()
	mix.CloseAudio()
	mix.Quit()
	img.Quit()
	sdl.Quit()
}

type Game struct {
	window         *sdl.Window
	renderer       *sdl.Renderer
	background     *sdl.Texture
	icon           *sdl.Surface
	fontSize       int
	fontColor      *sdl.Color
	text           *sdl.Texture
	textRect       *sdl.Rect
	textVelocity   int
	textXVelocity  int
	textYVelocity  int
	sprite         *sdl.Texture
	spriteRect     *sdl.Rect
	spriteVelocity int
	chunkGo        *mix.Chunk
	chunkSDL       *mix.Chunk
	music          *mix.Music
}

func NewGame() *Game {
	g := Game{}
	err := g.Init()
	if err != nil {
		panic(err)
	}
	return &g
}

func (g *Game) Init() error {
	var err error

	g.fontSize = 80
	g.fontColor = &sdl.Color{R: 255, G: 255, B: 255, A: 255}
	g.spriteVelocity = 10
	g.textVelocity = 2
	g.textXVelocity = g.textVelocity
	g.textYVelocity = g.textVelocity

	g.window, err = sdl.CreateWindow(windowTitle, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, windowWidth, windowHeight, 0) //sdl.WINDOWEVENT_SHOWN)
	if err != nil {
		return fmt.Errorf("Error creating window: %v", err)
	}

	g.renderer, err = sdl.CreateRenderer(g.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return fmt.Errorf("Error creating renderer: %v", err)
	}

	g.background, err = img.LoadTexture(g.renderer, "images/background.png")
	if err != nil {
		return fmt.Errorf("Error loading background image: %v", err)
	}

	g.icon, err = img.Load("images/Go-logo.png")
	if err != nil {
		return fmt.Errorf("Error loading icon image: %v", err)
	}
	g.window.SetIcon(g.icon)

	font, err := ttf.OpenFont("fonts/freesansbold.ttf", g.fontSize)
	if err != nil {
		return fmt.Errorf("Error loading font: %v", err)
	}
	textSurface, err := font.RenderUTF8Blended(windowTitle, *g.fontColor)
	if err != nil {
		return fmt.Errorf("Error creating font surface: %v", err)
	}
	defer textSurface.Free()

	g.text, err = g.renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		return fmt.Errorf("Error creating font texture: %v", err)
	}
	g.textRect = &sdl.Rect{X: (windowWidth - textSurface.W) / 2, Y: (windowHeight - textSurface.H) / 2, W: textSurface.W, H: textSurface.H}

	g.sprite, err = img.LoadTexture(g.renderer, "images/Go-logo.png")
	if err != nil {
		return fmt.Errorf("Error loading sprite image: %v", err)
	}
	g.spriteRect = &sdl.Rect{X: 0, Y: 0, W: spriteWidth, H: spriteHeight}

	err = mix.OpenAudio(mix.DEFAULT_FREQUENCY, mix.DEFAULT_FORMAT, mix.DEFAULT_CHANNELS, mix.DEFAULT_CHUNKSIZE)
	if err != nil {
		return fmt.Errorf("Error initializing SDL_mixer audio: %v", err)
	}

	g.chunkGo, err = mix.LoadWAV("sounds/Go.ogg")
	if err != nil {
		return fmt.Errorf("Error loading sound chunk: %v", err)
	}

	g.chunkSDL, err = mix.LoadWAV("sounds/SDL.ogg")
	if err != nil {
		return fmt.Errorf("Error loading sound chunk: %v", err)
	}

	g.music, err = mix.LoadMUS("music/freesoftwaresong-8bit.ogg")
	if err != nil {
		return fmt.Errorf("Error loading music: %v", err)
	}

	return nil
}

func (g *Game) Close() {
	if g == nil {
		return
	}

	mix.HaltMusic()
	mix.HaltChannel(-1)

	if g.window != nil {
		g.window.Destroy()
	}
	if g.renderer != nil {
		g.renderer.Destroy()
	}

	if g.background != nil {
		g.background.Destroy()
	}
	if g.icon != nil {
		g.icon.Free()
	}
	if g.textRect != nil {
		g.textRect = nil
	}
	if g.text != nil {
		g.text.Destroy()
	}
	if g.sprite != nil {
		g.sprite.Destroy()
	}
	if g.chunkGo != nil {
		g.chunkGo.Free()
	}
	if g.chunkSDL != nil {
		g.chunkSDL.Free()
	}
	if g.music != nil {
		g.music.Free()
	}
}

func (g *Game) Run() {
	g.music.Play(-1)

	fmt.Printf("%+v\n", g.spriteRect)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	go func(ticker *time.Ticker) {
		for range ticker.C {
			g.randColor()
		}
	}(ticker)

	for {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {

			switch e := event.(type) {
			case *sdl.QuitEvent:
				return
			case *sdl.KeyboardEvent:
				if e.Keysym.Sym == sdl.K_ESCAPE && e.Type == sdl.KEYDOWN {
					//fmt.Printf("Pressed %+v\n", e)
					return
				}
				if e.Keysym.Sym == sdl.K_SPACE && e.Type == sdl.KEYDOWN {
					g.chunkGo.Play(-1, 0)
					g.randColor()
				}
				if e.Keysym.Sym == sdl.K_m && e.Type == sdl.KEYDOWN {
					g.pauseUnpauseMusic()
				}
			}
		}

		keyboard := sdl.GetKeyboardState()
		if keyboard[sdl.SCANCODE_UP] != 0 || keyboard[sdl.SCANCODE_DOWN] != 0 || keyboard[sdl.SCANCODE_LEFT] != 0 || keyboard[sdl.SCANCODE_RIGHT] != 0 || keyboard[sdl.SCANCODE_W] != 0 || keyboard[sdl.SCANCODE_A] != 0 || keyboard[sdl.SCANCODE_S] != 0 || keyboard[sdl.SCANCODE_D] != 0 {
			g.moveSprite(keyboard)
		}
		// g.randColor() // Uncomment to change color every frame, gives seizures
		g.renderer.Clear()
		g.renderer.Copy(g.background, nil, nil)

		g.moveText()
		g.renderer.Copy(g.text, nil, g.textRect)
		g.renderer.Copy(g.sprite, nil, g.spriteRect)
		g.renderer.Present()

		sdl.Delay(20)
	}
}

func (g *Game) pauseUnpauseMusic() {
	if mix.PlayingMusic() {
		if mix.PausedMusic() {
			mix.ResumeMusic()
		} else {
			mix.PauseMusic()
		}
	}
}

func (g *Game) moveSprite(keyboard []uint8) {
	if keyboard[sdl.SCANCODE_UP] != 0 || keyboard[sdl.SCANCODE_W] != 0 {
		if g.spriteRect.Y >= 0 && g.spriteRect.Y-int32(g.spriteVelocity) >= 0 {
			g.spriteRect.Y -= int32(g.spriteVelocity)
		}
	}
	if keyboard[sdl.SCANCODE_DOWN] != 0 || keyboard[sdl.SCANCODE_S] != 0 {
		if g.spriteRect.Y+g.spriteRect.H < windowHeight && g.spriteRect.Y+g.spriteRect.H+int32(g.spriteVelocity) <= windowHeight {
			g.spriteRect.Y += int32(g.spriteVelocity)
		}
	}
	if keyboard[sdl.SCANCODE_LEFT] != 0 || keyboard[sdl.SCANCODE_A] != 0 {
		if g.spriteRect.X > 0 && g.spriteRect.X-int32(g.spriteVelocity) >= 0 {
			g.spriteRect.X -= int32(g.spriteVelocity)
		}
	}
	if keyboard[sdl.SCANCODE_RIGHT] != 0 || keyboard[sdl.SCANCODE_D] != 0 {
		if g.spriteRect.X+g.spriteRect.W < windowWidth && g.spriteRect.X+g.spriteRect.W+int32(g.spriteVelocity) <= windowWidth {
			g.spriteRect.X += int32(g.spriteVelocity)
		}
	}
	fmt.Printf("%+v\n", g.spriteRect)
}

func (g *Game) moveText() {
	g.textRect.X += int32(g.textXVelocity)
	g.textRect.Y += int32(g.textYVelocity)

	if g.textRect.X <= 0 || g.textRect.X+g.textRect.W >= windowWidth {
		g.textXVelocity = -g.textXVelocity
		g.chunkSDL.Play(-1, 0)
	}
	if g.textRect.Y <= 0 || g.textRect.Y+g.textRect.H >= windowHeight {
		g.textYVelocity = -g.textYVelocity
		g.chunkSDL.Play(-1, 0)
	}
}

func (g *Game) randColor() error {
	g.renderer.SetDrawColor(uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256)), 0)
	return nil
}

func main() {
	err := initSDL()
	if err != nil {
		panic(err)
	}
	defer closeSDL()

	g := NewGame()
	defer g.Close()

	g.Run()
}
