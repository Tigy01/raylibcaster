package rayrenderer

import (
	"image"
	"image/color"
	"log"
	"math"
	"raylibcaster/internal/levelmap"
	"raylibcaster/internal/player"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var RenderRes rl.Vector2
var currentFrame []color.RGBA

func DrawRays3D(renderTexture rl.RenderTexture2D, p player.Player, screenResolution rl.Vector2) {
	RenderRes = screenResolution //rl.Vector2Scale(screenRes, 1/RESOLUTION)

	distToPlane := calcDistanceToViewPlane(float64(p.FOV))
	currentFrame = make([]color.RGBA, int(RenderRes.X*RenderRes.Y))

	doneChan := make(chan bool, int(RenderRes.X))

	for r := range int(RenderRes.X) {
		rayAngle := calcNextViewAngle(distToPlane, float64(r)) + p.Angle
		if rayAngle < 0 {
			rayAngle += 2 * math.Pi
		}
		if rayAngle > 2*math.Pi {
			rayAngle -= 2 * math.Pi
		}

		go drawRayWall3D(p, rayAngle, r, doneChan)
	}
	for range cap(doneChan) {
		<-doneChan
	}
	rl.UpdateTexture(renderTexture.Texture, currentFrame)
}

// Uses trig to find the appropriate distance from the player to a plane that is the width of the
// resolution of the screen given a fieldOfView. Essentially, screenRes/2 is the opposite measure becasue
// when split in half the view cone is a right triangle. This finds the adjacent angle.
func calcDistanceToViewPlane(fieldOfView float64) float64 {
	return (float64(RenderRes.X) / 2) / math.Tan(rl.Deg2rad*fieldOfView/2)
}

// This function takes the calculated distToPlane (adjacent) and uses the rayNumber as a pixel offset of the
// screenRes (opposite) to calculate the angle of the ray that we will cast.
//
// 0.5 is subtracted because, on even resolutions ex: 1024, rayNumber - screenRes.X is the number of
// the pixel collum we are finding screen over the range 0 - screenRes.X. Because it steps by one,
// subtracting 0.5 results in the rays shooting down the middle of the pixel
func calcNextViewAngle(distToPlane float64, rayNumber float64) float64 {
	return math.Atan2(float64(rayNumber)-0.5-float64(RenderRes.X)/2, distToPlane)
}

func drawRayWall3D(p player.Player, rayAngle float64, rayNumber int, doneChan chan bool) {
	hRay, _, minRay := castRayFromPosition(p.Position, rayAngle)
	rayLen := rl.Vector2Length(rl.Vector2Subtract(minRay, p.Position))

	angleDelta := p.Angle - rayAngle
	if angleDelta < 0 {
		angleDelta += 2 * math.Pi
	} else if angleDelta > 2*math.Pi {
		angleDelta -= 2 * math.Pi
	}

	rayLen *= float32(math.Cos(angleDelta)) // fix warping

	lineH := float32(levelmap.MapScale) * RenderRes.Y / rayLen

	cellType := levelmap.GetMapCellFromPosition(minRay)

	var cellImage image.Image
	if cI, ok := levelmap.Images.Load(cellType); !ok {
		log.Fatalf("invalid image id %v", cellType)
		return
	} else {
		if cellImage, ok = cI.(image.Image); !ok {
			log.Fatalf("Image not loaded properly at id:%v", cellType)
		}
	}

	ty_step := float32(cellImage.Bounds().Dy()) / lineH

	var textureYOff float32 = 0

	if lineH > RenderRes.Y {
		textureYOff = (lineH - RenderRes.Y) / 2.0
		lineH = RenderRes.Y
	}

	lineO := RenderRes.Y/2 - lineH/2

	var textureX int
	if rl.Vector2Equals(minRay, hRay) {
		textureX = int(minRay.X) % cellImage.Bounds().Dx()
	} else {
		textureX = int(minRay.Y) % cellImage.Bounds().Dx()
	}

	if levelmap.IsOnMap(minRay) {
		MapTextureToFrame(cellImage, textureX, ty_step*textureYOff, ty_step, rayNumber, lineO, lineH)
	}
	doneChan <- true
}

func MapTextureToFrame(cellImage image.Image, textureX int, textureY, step float32, x int, lineO, lineH float32) {
	oldTy := 0
	var rgba color.RGBA
	for y := float32(0); y < lineH; y++ {
		if oldTy != int(textureY) { //prevents reatlasing the texture every pixel
			c := cellImage.At(textureX, int(textureY))
			r, g, b, a := c.RGBA()
			rgba = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
			oldTy = int(textureY)
		}
		DrawColorToFrame(x, int(y+lineO), rgba)
		textureY += step
	}
}

func DrawColorToFrame(x, y int, color color.RGBA) {
	index := y*int(RenderRes.X) + x
	currentFrame[index] = color
}

func castRayFromPosition(position rl.Vector2, angle float64) (hRay, vRay, minRay rl.Vector2) {
	hRay = horizontalChecks(position, angle)
	vRay = verticalChecks(position, angle)

	offsetHRay := rl.Vector2Subtract(hRay, position)
	offsetVRay := rl.Vector2Subtract(vRay, position)

	minRay = getShortestRay(offsetHRay, offsetVRay)
	minRay = rl.Vector2Add(minRay, position)
	return
}

func getShortestRay(ray1, ray2 rl.Vector2) rl.Vector2 {
	if rl.Vector2Length(ray1) < rl.Vector2Length(ray2) {
		return ray1
	}
	return ray2
}

func horizontalChecks(playerPos rl.Vector2, rayAngle float64) (rPos rl.Vector2) {
	var dof int
	var xOffset, yOffset float64
	var rayX, rayY float64
	aTan := -1.0 / math.Tan(rayAngle)

	rayY = float64((int(playerPos.Y) >> 6) << 6)
	yOffset = float64(levelmap.MapScale)
	if rayAngle > math.Pi { //up
		rayY -= 0.0001
		yOffset *= -1
	}
	if rayAngle < math.Pi { //down
		rayY += float64(levelmap.MapScale)
	}
	rayX = (float64(playerPos.Y)-rayY)*aTan + float64(playerPos.X)
	xOffset = -1 * yOffset * aTan

	if rayAngle == 0 || rayAngle == math.Pi || rayAngle == 2*math.Pi {
		rayX = float64(playerPos.X)
		rayY = float64(playerPos.Y)
		dof = 8
	}

	for dof < 8 {
		if levelmap.GetMapCellFromPosition(rl.NewVector2(float32(rayX), float32(rayY))) > 0 {
			dof = 8
		} else {
			rayX += xOffset
			rayY += yOffset
			dof += 1
		}
	}
	return rl.NewVector2(float32(rayX), float32(rayY))
}

func verticalChecks(position rl.Vector2, rayAngle float64) (rPos rl.Vector2) {
	var xOffset, yOffset float64
	var rayX, rayY float64
	var dof int
	nTan := -math.Tan(rayAngle)

	rayX = float64((int(position.X) >> 6) << 6)
	xOffset = float64(levelmap.MapScale)
	if rayAngle > math.Pi/2 && rayAngle < 3*math.Pi/2 {
		rayX -= 0.0001
		xOffset *= -1
	}
	if rayAngle < math.Pi/2 || rayAngle > 3*math.Pi/2 {
		rayX += float64(levelmap.MapScale)
	}

	rayY = (float64(position.X)-rayX)*nTan + float64(position.Y)
	yOffset = -1 * xOffset * nTan

	if rayAngle == math.Pi/2 || rayAngle == 3*math.Pi/2 {
		rayX = float64(position.X)
		rayY = float64(position.Y)
		dof = 8
	}

	for dof < 8 {
		if levelmap.GetMapCellFromPosition(rl.NewVector2(float32(rayX), float32(rayY))) > 0 {
			dof = 8
		} else {
			rayX += xOffset
			rayY += yOffset
			dof += 1
		}
	}

	return rl.NewVector2(float32(rayX), float32(rayY))
}

func drawRayLine(startPos, endPos rl.Vector2, color color.RGBA) {
	if endPos.X > RenderRes.X {
		xDiff := endPos.X - startPos.X
		angle := math.Atan2(float64(endPos.Y-startPos.Y), float64(xDiff))
		yCoord := (RenderRes.X - startPos.X) * float32(math.Tan(angle))
		endPos.Y = yCoord + startPos.Y
		endPos.X = RenderRes.X
	}
	rl.DrawLineV(
		startPos,
		endPos,
		color,
	)
}
