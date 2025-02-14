package rayrenderer

import (
	"image/color"
	"math"
	"raylibcaster/internal/levelmap"
	"raylibcaster/internal/player"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const RESOLUTION = 1.0
const RAYOFFSET = rl.Deg2rad / RESOLUTION

var lineWidth int = RESOLUTION
var screenRes rl.Vector2
var RenderRes rl.Vector2


func DrawRays3D(p player.Player, screenResolution rl.Vector2) {
	screenRes = screenResolution
	RenderRes = rl.Vector2Scale(screenRes, 1/RESOLUTION)

	distToPlane := calcDistanceToViewPlane(float64(p.FOV))

	for r := range int(RenderRes.X) {
		rayAngle := calcNextViewAngle(distToPlane, float64(r)) + p.Angle
		if rayAngle < 0 {
			rayAngle += 2 * math.Pi
		}
		if rayAngle > 2*math.Pi {
			rayAngle -= 2 * math.Pi
		}

		drawRayWall3D(p, rayAngle, r)
	}
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

func drawRayWall3D(p player.Player, rayAngle float64, rayNumber int) {
	hRay, _, minRay := castRayFromPosition(p.Position, rayAngle)
	rayLen := rl.Vector2Length(rl.Vector2Subtract(minRay, p.Position))

	angleDelta := p.Angle - rayAngle
	if angleDelta < 0 {
		angleDelta += 2 * math.Pi
	} else if angleDelta > 2*math.Pi {
		angleDelta -= 2 * math.Pi
	}

	rayLen *= float32(math.Cos(angleDelta)) // fix warping

	lineH := float32(levelmap.MapScale) * screenRes.Y / rayLen

	lineO := screenRes.Y/2 - lineH/2

	texId := int(levelmap.GetMapCellFromPosition(minRay))

	var x int
	if rl.Vector2Equals(minRay, hRay) {
		x = int(minRay.X) * levelmap.Images[texId].Bounds().Dx() / int(levelmap.MapScale) % levelmap.Images[texId].Bounds().Dx()
		if rayAngle < math.Pi {
			x = levelmap.Images[texId].Bounds().Dx() - x - 1
		}
	} else {
		x = int(minRay.Y) * levelmap.Images[texId].Bounds().Dx() / int(levelmap.MapScale) % levelmap.Images[texId].Bounds().Dx()
		if rayAngle > math.Pi/2 && rayAngle < 3*math.Pi/2 {
			x = levelmap.Images[texId].Bounds().Dx() - x - 1
		}
	}

	screenPosition := rl.NewVector2(
		float32(rayNumber*lineWidth),
		lineO,
	)
	lineScale := rl.NewVector2(float32(lineWidth), lineH)

	if levelmap.IsOnMap(minRay) {
		drawLineOfTexture(x, texId, screenPosition, lineScale)
	}
}

func drawLineOfTexture(x, texId int, screenPosition, lineScale rl.Vector2) {
	pixelSizeY := lineScale.Y / float32(levelmap.Images[texId].Bounds().Dx())
	for y := range levelmap.Images[texId].Bounds().Dy() {

		pixelColor := levelmap.Images[texId].At(x, y)
		r, g, b, a := pixelColor.RGBA()
		pixelRGBA := color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

		pixelPosition := rl.NewVector2(screenPosition.X, screenPosition.Y+float32(y)*pixelSizeY)
		pixelScale := rl.NewVector2(lineScale.X, pixelSizeY)
		rl.DrawRectangleV(pixelPosition, pixelScale, pixelRGBA)
	}
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
