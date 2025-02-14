package player

import (
	"image/color"
	"math"
	"raylibcaster/internal/levelmap"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var playerScale int32 = 8

type Player struct {
	Position  rl.Vector2
	Velocity  rl.Vector2
	Speed     float32
	TurnSpeed float32
	Angle     float64
	FOV       int
}

func Init(position rl.Vector2, movementSpeed, turnSpeed float32) *Player {
	p := &Player{
		Position:  position,
		Velocity:  rl.NewVector2(0, 0),
		Angle:     2 * math.Pi,
		FOV:       90,
		TurnSpeed: turnSpeed * rl.Deg2rad,
		Speed:     movementSpeed,
	}
	p.calculateVelocity(1.0 / 60.0)
	return p
}

func (p Player) Draw() {
	rl.DrawRectangle(
		int32(p.Position.X-float32(playerScale)/2),
		int32(p.Position.Y-float32(playerScale)/2),
		playerScale,
		playerScale,
		color.RGBA{255, 255, 0, 255},
	)
	rl.DrawLine(
		int32(p.Position.X),
		int32(p.Position.Y),
		int32(p.Position.X+p.Velocity.X*5),
		int32(p.Position.Y+p.Velocity.Y*5),
		color.RGBA{255, 255, 0, 255},
	)
}

var UP = rl.NewVector2(0, -1)
var DOWN = rl.NewVector2(0, 1)
var LEFT = rl.NewVector2(-1, 0)
var RIGHT = rl.NewVector2(1, 0)

func (p *Player) Process() {
}

func (p *Player) Input() {
	if rl.IsKeyDown(rl.KeyLeft) {
		p.Angle -= float64(p.TurnSpeed * rl.GetFrameTime())
		if p.Angle < 0 {
			p.Angle += 2 * math.Pi
		}
	}
	if rl.IsKeyDown(rl.KeyRight) {
		p.Angle += float64(p.TurnSpeed * rl.GetFrameTime())
		if p.Angle > 2*math.Pi {
			p.Angle -= 2 * math.Pi
		}
	}
	if p.Angle < 0.00001 {
		p.Angle = 0
	}

	if rl.IsKeyDown(rl.KeyA) {
        p.calculateVelocity(rl.GetFrameTime())

		p.Velocity = rl.NewVector2(p.Velocity.Y, -p.Velocity.X)
		p.moveAndCollide()
		p.Velocity = rl.NewVector2(-p.Velocity.Y, p.Velocity.X)
	} else if rl.IsKeyDown(rl.KeyD) {
        p.calculateVelocity(rl.GetFrameTime())

		p.Velocity = rl.NewVector2(-p.Velocity.Y, p.Velocity.X)
		p.moveAndCollide()
		p.Velocity = rl.NewVector2(p.Velocity.Y, -p.Velocity.X)
	}

	if rl.IsKeyDown(rl.KeyW) {
		p.calculateVelocity(rl.GetFrameTime())
		p.moveAndCollide()
	}
	if rl.IsKeyDown(rl.KeyS) {
        p.calculateVelocity(rl.GetFrameTime())
		p.Velocity = rl.Vector2Scale(p.Velocity, -1)

		p.moveAndCollide()

		p.Velocity = rl.Vector2Scale(p.Velocity, -1)
	}
	if rl.IsKeyPressed(rl.KeyF) {

		if rl.IsKeyDown(rl.KeyLeftShift) {

			p.FOV += 1
		} else {
			p.FOV -= 1
		}
	}
}

func (p *Player) moveAndCollide() {
	nextPos := rl.Vector2Add(p.Position, p.Velocity)
	if levelmap.GetMapCellFromPosition(nextPos) == 0 {
		p.Position = nextPos
		return
	}
	nextXPos := rl.Vector2Add(p.Position, rl.NewVector2(p.Velocity.X, 0))
	if levelmap.GetMapCellFromPosition(nextXPos) == 0 {
		p.Position = nextXPos
		return
	}
	nextYPos := rl.Vector2Add(p.Position, rl.NewVector2(0, p.Velocity.Y))
	if levelmap.GetMapCellFromPosition(nextYPos) == 0 {
		p.Position = nextYPos
		return
	}
}

func (p *Player) calculateVelocity(delta float32) {
	p.Velocity.X = float32(math.Cos(float64(p.Angle)))
	p.Velocity.Y = float32(math.Sin(float64(p.Angle)))
	p.Velocity = rl.Vector2Scale(rl.Vector2Normalize(p.Velocity), p.Speed*delta)
}
