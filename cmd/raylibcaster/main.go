package main

import (
	"fmt"
	"image/color"
	"raylibcaster/internal/levelmap"
	"raylibcaster/internal/player"
	"raylibcaster/internal/rayrenderer"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var resolution = rl.NewVector2(1024, 512)

func main() {
	rl.SetTraceLogLevel(rl.LogError)
	rl.InitWindow(int32(resolution.X), int32(resolution.Y), "raycaster")
	rl.SetTargetFPS(120)
	rl.SetWindowState(rl.FlagWindowResizable)
	rl.SetConfigFlags(rl.FlagVsyncHint)

	p := player.Init(rl.NewVector2(300, 300), 150, 90)

	levelmap.LoadImage("./assets/wall32.png", 2, true)
	levelmap.LoadImage("./assets/brick.png", 1, true)

	averageFPS := float64(0)
	frameCount := float64(0)

	for !rl.WindowShouldClose() {
		resolution = rl.NewVector2(1024, 512)
		renderTex := rl.LoadRenderTexture(int32(resolution.X), int32(resolution.Y))
		resolution = rl.NewVector2(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))

		p.Input()
		p.Process()

		rl.BeginTextureMode(renderTex)

		rayrenderer.DrawRays3D(renderTex, *p)

		rl.EndTextureMode()

		rl.BeginDrawing()

		rl.DrawRectangle(
			0,
			0,
			int32(resolution.X),
			int32(resolution.Y)/2,
			rl.SkyBlue,
		) // floor

		rl.DrawRectangle(
			0,
			int32(resolution.Y)/2,
			int32(resolution.X),
			int32(resolution.Y)/2,
			rl.Brown,
		) // floor

		drawRenderTexture(renderTex)

		rl.DrawText(
			fmt.Sprintf("average: %0.0f\ncurrent: %d", averageFPS, rl.GetFPS()),
			0, 0, 32, rl.Black,
		)
		rl.EndDrawing()

		frameCount += 1

		averageFPS = averageFPS*(frameCount-1)/frameCount + float64(rl.GetFPS())/frameCount
		rl.UnloadRenderTexture(renderTex)
	}

	rl.CloseWindow()
}

func drawRenderTexture(renderTex rl.RenderTexture2D) {
	rl.DrawTexturePro(renderTex.Texture, rl.Rectangle{
		X:      0,
		Y:      float32(renderTex.Texture.Height),
		Width:  float32(renderTex.Texture.Width),
		Height: float32(-renderTex.Texture.Height),
	},
		rl.Rectangle{
			X:      0,
			Y:      0,
			Width:  float32(rl.GetScreenWidth()),
			Height: float32(rl.GetScreenHeight()),
		},
		rl.Vector2{
			X: 0,
			Y: 0,
		},
		0,
		color.RGBA{255, 255, 255, 255})
}
