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
	rl.SetWindowState(rl.FlagWindowResizable)
	rl.SetConfigFlags(rl.FlagVsyncHint)

	p := player.Init(rl.NewVector2(300, 300), 150, 90)

	levelmap.LoadWallImage("./assets/wall32.png", 1)
	levelmap.LoadWallImage("./assets/brick.png", 2)

	for !rl.WindowShouldClose() {
		p.Input()
		p.Process()

		resolution = rl.NewVector2(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()))
		rl.BeginDrawing()

		rl.ClearBackground(color.RGBA{77, 77, 77, 255})

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
		rayrenderer.DrawRays3D(*p, resolution)

		rl.DrawText(fmt.Sprint(rl.GetFPS()), 0, 0, 32, rl.Yellow)
		rl.EndDrawing()
	}
	rl.CloseWindow()
}
