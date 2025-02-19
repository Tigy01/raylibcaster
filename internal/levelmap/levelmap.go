package levelmap

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var MapScale int32 = 64 
var MapX int32 = 12
var MapY int32 = 12

type MapCell struct {
	Texture image.Image
	IsWall  bool
}

var Map = [][]int{
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 2, 1, 1, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 2, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	{1, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1},
}

var Cells sync.Map

func DrawMap() {
	for y := range MapY {
		for x := range MapX {
			position := rl.NewVector2(float32(x*MapScale), float32(y*MapScale))
			if Map[y][x] == 1 {
				drawPixel(position, float32(MapScale-1), color.RGBA{255, 255, 255, 255})
			} else {
				drawPixel(position, float32(MapScale-1), color.RGBA{0, 0, 0, 255})
			}
		}
	}
}

func drawPixel(position rl.Vector2, size float32, color color.RGBA) {
	rl.DrawRectangleV(
		position,
		rl.NewVector2(size, size),
		color,
	)
}

func GetMapSpaceCoordinate(pos rl.Vector2) rl.Vector2 {
	mx := pos.X / float32(MapScale)
	my := pos.Y / float32(MapScale)

	if int(my) >= len(Map) || int(mx) >= len(Map[0]) || mx < 0 || my < 0 {
		return rl.NewVector2(-1, -1)
	}

	return rl.NewVector2(mx, my)
}

func Vector2IsInRange(vec, low, high rl.Vector2) bool {
	if vec.X >= low.X && vec.Y >= low.Y {
		if vec.Y < high.Y && vec.X < high.X {
			return true
		}
	}
	return false
}

func GetMapCellFromPosition(position rl.Vector2) *MapCell {
	mapIndex := GetMapSpaceCoordinate(position)
	if Vector2IsInRange(mapIndex, rl.NewVector2(0, 0), rl.NewVector2(float32(len(Map[0])), float32(len(Map)))) {
		rawCell, ok := Cells.Load(Map[int(mapIndex.Y)][int(mapIndex.X)])
		if ok {
			if cell, ok := rawCell.(MapCell); ok {
				return &cell
			}
		}
	}
	return nil
}

func LoadImage(path string, id int, isWall bool) (err error) {
	wallFile, err := os.Open(path)
	if err != nil {
		return err
	}
	wallImage, err := png.Decode(wallFile)
	if err != nil {
		return err
	}

	cell := MapCell{
		Texture: wallImage,
		IsWall:  isWall,
	}

	Cells.LoadOrStore(id, cell)
	wallFile.Close()
	return nil
}

func IsOnMap(ray rl.Vector2) bool {
	mapPos := GetMapSpaceCoordinate(ray)
	return Vector2IsInRange(mapPos, rl.NewVector2(0, 0), rl.NewVector2(float32(len(Map[0])), float32(len(Map))))
}
