package levelmap

import (
	"image"
	"image/color"
	"image/png"
	"os"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var MapScale int32 = 64
var MapX int32 = 8
var MapY int32 = 8
var Map = []int{
	1, 2, 2, 2, 2, 2, 2, 1,
	1, 0, 0, 0, 0, 0, 0, 1,
	1, 0, 0, 0, 0, 0, 0, 1,
	1, 0, 0, 0, 2, 1, 0, 1,
	1, 0, 0, 0, 0, 0, 0, 1,
	1, 0, 0, 0, 0, 0, 0, 1,
	1, 0, 0, 0, 0, 0, 0, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
}

var Images = map[int]image.Image{}

func DrawMap() {
	for y := range MapY {
		for x := range MapX {
			position := rl.NewVector2(float32(x*MapScale), float32(y*MapScale))
			if Map[y*MapX+x] == 1 {
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

func GetMapSpaceCoordinate(pos rl.Vector2) int {
	mx := int(pos.X / float32(MapScale))
	my := int(pos.Y / float32(MapScale))

	if my >= int(MapY) || mx >= int(MapX) || mx < 0 || my < 0 {
		return -1
	}

	return int(my*int(MapX) + mx)
}

func GetMapCellFromPosition(position rl.Vector2) int {
	mapPos := GetMapSpaceCoordinate(position)
	if mapPos >= 0 && mapPos < len(Map) {
		return Map[mapPos]
	} else {
		return -1000
	}
}

func LoadWallImage(path string, id int) (err error) {
	wallFile, err := os.Open(path)
	if err != nil {
		return err
	}
	wallImage, err := png.Decode(wallFile)
	if err != nil {
		return err
	}
	Images[id] = wallImage
	wallFile.Close()
	return nil
}

func IsOnMap(ray rl.Vector2) bool {
	if ray.X < 0 || ray.X > float32(MapScale*MapX) {
		return false
	}
	if ray.Y < 0 || ray.Y > float32(MapScale*MapX) {
		return false
	}
	return true
}
