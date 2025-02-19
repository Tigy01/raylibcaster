package rayrenderer

import (
	"image"
	"image/color"
	"math"
	"raylibcaster/internal/levelmap"
	"raylibcaster/internal/player"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var RenderRes rl.Vector2
var currentFrame []color.RGBA

func DrawRays3D(renderTexture rl.RenderTexture2D, p player.Player) {
	RenderRes = rl.NewVector2(float32(renderTexture.Texture.Width), float32(renderTexture.Texture.Height)) //rl.Vector2Scale(screenRes, 1/RESOLUTION)

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
	defer func() { doneChan <- true }()

	hRay, _, minRay := castRayFromPosition(p.Position, rayAngle)
	rayLen := rl.Vector2Length(rl.Vector2Subtract(minRay, p.Position))

	angleDelta := p.Angle - rayAngle
	if angleDelta < 0 {
		angleDelta += 2 * math.Pi
	} else if angleDelta > 2*math.Pi {
		angleDelta -= 2 * math.Pi
	}

	if levelmap.IsOnMap(minRay) {
		MapTextureToFrame(rayNumber, minRay, hRay, rayAngle, angleDelta, rayLen)
	}
}

func MapTextureToFrame(pixelX int, minRay, hRay rl.Vector2, rayAngle, angleDelta float64, rayLen float32) {
	shade := float64(250 / rayLen)
	if shade > 1 {
		shade = 1
	}

	rayLen *= float32(math.Cos(angleDelta)) // fix warping

	lineH := float32(levelmap.MapScale) * RenderRes.Y / rayLen

	var cell *levelmap.MapCell = levelmap.GetMapCellFromPosition(minRay)
	if cell == nil || !cell.IsWall {
		return
	}

	//	cellImage := cell
	var cellImage image.Image = cell.Texture

	textureY, ty_step := getTextureY(cellImage, lineH)
	textureX := getTextureX(minRay, hRay, rayAngle, cellImage)

	lineH = rl.Clamp(lineH, 0, RenderRes.Y)
	lineO := (RenderRes.Y - lineH) / 2.0
	oldTy := -1

	var rgba color.RGBA
	for y := float32(0); y < lineH; y++ {
		if oldTy != int(textureY) { //prevents reatlasing the texture every pixel
			rgba = getRGBA(cellImage, textureX, int(textureY))
			rgba = changeBrightness(rgba, shade)
			oldTy = int(textureY)
		}
		DrawColorToFrame(pixelX, int(y+lineO), rgba)
		textureY += ty_step
	}
}

func changeBrightness(rgba color.RGBA, shade float64) color.RGBA {
	rgba.R = uint8(float64(uint32(rgba.R)) * shade)
	rgba.G = uint8(float64(uint32(rgba.G)) * shade)
	rgba.B = uint8(float64(uint32(rgba.B)) * shade)
	return rgba
}

func getTextureY(cellImage image.Image, lineH float32) (textureY, ty_step float32) {
	ty_step = float32(cellImage.Bounds().Dy()) / lineH
	var textureYOff float32 = 0
	if lineH > RenderRes.Y {
		textureYOff = (lineH - RenderRes.Y) / 2.0
	}
	textureY = ty_step * textureYOff
	return
}

func getTextureX(minRay, hRay rl.Vector2, rayAngle float64, cellImage image.Image) (textureX int) {
	if rl.Vector2Equals(minRay, hRay) {
		textureX = (int(minRay.X) * cellImage.Bounds().Dx() / int(levelmap.MapScale)) % cellImage.Bounds().Dx()
		if rayAngle < math.Pi {
			textureX = cellImage.Bounds().Dx() - textureX - 1
		}
	} else {
		textureX = (int(minRay.Y) * cellImage.Bounds().Dx() / int(levelmap.MapScale)) % cellImage.Bounds().Dx()
		if rayAngle > math.Pi/2 && rayAngle < 3*math.Pi/2 {
			textureX = cellImage.Bounds().Dx() - textureX - 1
		}
	}
	return
}

func getRGBA(cellImage image.Image, x, y int) color.RGBA {
	c := cellImage.At(x, y)
	r, g, b, a := c.RGBA()
	rgba := color.RGBA{
		uint8(r),
		uint8(g),
		uint8(b),
		uint8(a),
	}
	return rgba
}

func DrawColorToFrame(x, y int, color color.RGBA) {
	y = int(RenderRes.Y-1) - y // flips y coordinate because its flipped in the render texture
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

const MAX_DOF = 12

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
		dof = MAX_DOF
	}

	for dof < MAX_DOF {
		if cell := levelmap.GetMapCellFromPosition(rl.NewVector2(float32(rayX), float32(rayY))); cell != nil && cell.IsWall {
			dof = MAX_DOF
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
		dof = MAX_DOF
	}

	for dof < MAX_DOF {
		if cell := levelmap.GetMapCellFromPosition(rl.NewVector2(float32(rayX), float32(rayY))); cell != nil && cell.IsWall {
			dof = MAX_DOF
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
