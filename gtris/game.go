package gtris

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

const (
	ScreenWidth  = 240
	ScreenHeight = 360
)

type Size struct {
	Width  uint
	Height uint
}

type GameState int

const (
	GameStateGameOver GameState = iota
	GameStatePlaying
)

// Button представляет круглую кнопку
type Button struct {
	X, Y    float64 // Центр кнопки
	Radius  float64 // Радиус
	Color   color.RGBA
	Pressed bool
	PressedLong bool
}

// NewButton создает новую кнопку
func NewButton(x, y, radius float64, color color.RGBA) *Button {
	return &Button{
		X:      x,
		Y:      y,
		Radius: radius,
		Color:  color,
	}
}

// Contains проверяет, находится ли точка внутри кнопки
func (b *Button) Contains(x, y int) bool {
	dx := float64(x) - b.X
	dy := float64(y) - b.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance <= b.Radius
}

// Update обновляет состояние кнопок
func (b *Button) Update(cursorX, cursorY int, isPressed bool) {
	if b.Contains(cursorX, cursorY) && isPressed {
		if b.Pressed {
			b.PressedLong = true
		}
		b.Pressed = true
	} else {
		b.Pressed = false
		b.PressedLong = false
	}
}

// Простой шейдер для рисования (требуется для vector.Path)
var shader *ebiten.Shader

func init() {
	// Компилируем простой шейдер
	var err error
	shader, err = ebiten.NewShader([]byte(`
		package main

		func Fragment(position vec4, texCoord vec2, color vec4) vec4 {
			return color
		}
	`))
	if err != nil {
		log.Fatal(err)
	}
}

// Draw рисует кнопку
func (b *Button) Draw(screen *ebiten.Image) {
	// Цвет кнопки (темнее при нажатии)
	btnColor := b.Color
	if b.Pressed {
		btnColor = color.RGBA{
			R: uint8(float32(b.Color.R) * 0.7),
			G: uint8(float32(b.Color.G) * 0.7),
			B: uint8(float32(b.Color.B) * 0.7),
			A: b.Color.A,
		}
	}

	// Рисуем заполненный круг (современный API)
	path := vector.Path{}
	path.Arc(float32(b.X), float32(b.Y), float32(b.Radius), 0, 2*math.Pi, vector.Clockwise)
	vertices, indices := path.AppendVerticesAndIndicesForFilling(nil, nil)

	// Применяем цвет ко всем вершинам
	for i := range vertices {
		vertices[i].ColorR = float32(btnColor.R) / 255
		vertices[i].ColorG = float32(btnColor.G) / 255
		vertices[i].ColorB = float32(btnColor.B) / 255
		vertices[i].ColorA = float32(btnColor.A) / 255
	}

	screen.DrawTrianglesShader(vertices, indices, shader, &ebiten.DrawTrianglesShaderOptions{
		Blend: ebiten.BlendCopy,
	})

	// Рисуем обводку
	strokePath := vector.Path{}
	strokePath.Arc(float32(b.X), float32(b.Y), float32(b.Radius), 0, 2*math.Pi, vector.Clockwise)
	strokeOp := &vector.StrokeOptions{}
	strokeOp.Width = 3
	strokeVertices, strokeIndices := strokePath.AppendVerticesAndIndicesForStroke(nil, nil, strokeOp) // 3 - толщина линии

	// Черный цвет для обводки
	for i := range strokeVertices {
		strokeVertices[i].ColorR = 0
		strokeVertices[i].ColorG = 0
		strokeVertices[i].ColorB = 0
		strokeVertices[i].ColorA = 1
	}

	screen.DrawTrianglesShader(strokeVertices, strokeIndices, shader, &ebiten.DrawTrianglesShaderOptions{
		Blend: ebiten.BlendSourceOver,
	})
}

type Game struct {
	dropTicks   uint
	elapsedDrop uint

	score       int
	state       GameState
	attractMode bool
	pieces      []*Piece

	nextPiece     *Piece
	currentPiece  *Piece
	piecePosition *Position

	gameZoneSize Size
	gameZone     [][]*ebiten.Image
	bgBlockImage *ebiten.Image

	txtFont font.Face

	input            Input
	inputAttractMode Input
	inputKeyboard    Input

	smallButtons []*Button
	largeButton  *Button

	screenWidth  int
	screenHeight int
}

func (g *Game) SetScreenWidth(screenWidth int) {
	g.screenWidth = screenWidth
}

func (g *Game) SetScreenHeight(screenHeight int) {
	g.screenHeight = screenHeight
}

func (g *Game) Start() {
	g.state = GameStatePlaying
	g.score = 0
	g.attractMode = true
	g.input = g.inputAttractMode
	g.elapsedDrop = 0

	g.gameZone = make([][]*ebiten.Image, g.gameZoneSize.Height)
	for y := range g.gameZone {
		g.gameZone[y] = make([]*ebiten.Image, g.gameZoneSize.Width)
	}

	g.fetchNextPiece()
}

func (g *Game) StartPlay() {
	g.Start()
	g.attractMode = false
	g.input = g.inputKeyboard
}

func (g *Game) Update() error {
	g.elapsedDrop += 1

	switch g.state {
	case GameStatePlaying:
		if g.elapsedDrop > g.dropTicks {
			g.processInput(keyDown)
			g.elapsedDrop = 0
			return nil
		}

		key := g.input.Read()
		if key != nil {
			g.processInput(*key)
		}
		tch := g.processTouchReturnKey()
		if tch != 0 {
			g.processInput(tch)
		}

		if g.attractMode && g.inputKeyboard.IsSpacePressed() {
			g.StartPlay()
		}

		if g.attractMode && tch == ebiten.KeySpace {
			g.StartPlay()
		}
	case GameStateGameOver:
		if g.input.IsSpacePressed() {
			if g.attractMode {
				g.Start()
			} else {
				g.StartPlay()
			}
		}

		tch := g.processTouchReturnKey()
		if tch == ebiten.KeySpace {
			if g.attractMode {
				g.Start()
			} else {
				g.StartPlay()
			}
		}
	}

	return nil
}

func (g *Game) processPiece() bool {
	g.transferPieceToGameZone()
	linesRemoved := g.checkForLines()
	g.updateScore(linesRemoved)
	g.fetchNextPiece()

	stopProcess := false
	deltaPos := Position{}
	if !g.insideGameZone(deltaPos) {
		g.state = GameStateGameOver
		stopProcess = true
	}

	return stopProcess
}

func (g *Game) processInput(key ebiten.Key) {
	if key == ebiten.KeyDown {
		deltaPos := Position{X: 0, Y: 1}
		if g.insideGameZone(deltaPos) {
			g.piecePosition.Add(deltaPos)
		} else {
			stopProcess := g.processPiece()
			if stopProcess {
				return
			}
		}
	}

	if key == ebiten.KeyLeft {
		deltaPos := Position{X: -1, Y: 0}
		if g.insideGameZone(deltaPos) {
			g.piecePosition.Add(deltaPos)
		}
	}

	if key == ebiten.KeyRight {
		deltaPos := Position{X: 1, Y: 0}
		if g.insideGameZone(deltaPos) {
			g.piecePosition.Add(deltaPos)
		}
	}

	if key == ebiten.KeySpace {
		newPiece := g.rotatePiece()
		if g.pieceInsideGameZone(newPiece, *g.piecePosition) {
			g.currentPiece = newPiece
		}
	}
}

func (g *Game) processTouchReturnKey() ebiten.Key {

	// Получаем позицию курсора и состояние мыши
	cursorX, cursorY := ebiten.CursorPosition()
	isMousePressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Проверяем touch input (для мобильных устройств)
	if touches := ebiten.AppendTouchIDs(nil); len(touches) > 0 {
		// Берем первое касание
		touchX, touchY := ebiten.TouchPosition(touches[0])
		cursorX, cursorY = touchX, touchY
		isMousePressed = true
	}
	// Обновляем состояние маленьких кнопок
	for _, btn := range g.smallButtons {
		btn.Update(cursorX, cursorY, isMousePressed)
	}

	// Обновляем состояние большой кнопки
	g.largeButton.Update(cursorX, cursorY, isMousePressed)

	// В методе Update после обновления кнопок:
	for i, btn := range g.smallButtons {
		if btn.Pressed {
			switch i {
			case 0:
				println("Нажата маленькая кнопка", 1)
				return ebiten.KeyLeft
			case 1:
				println("Нажата маленькая кнопка", 2)
				return ebiten.KeyUp
			case 2:
				println("Нажата маленькая кнопка", 3)
				return ebiten.KeyRight
			case 3:
				println("Нажата маленькая кнопка", 4)
				return ebiten.KeyDown

			}
		}
	}
	if g.largeButton.Pressed {
		println("Нажата большая кнопка!")
		if g.largeButton.PressedLong {
			return 0
		}
		return ebiten.KeySpace
	}
	return 0
}

func (g *Game) drawText(screen *ebiten.Image, gameZonePos *Position) {
	boardBlock := g.bgBlockImage.Bounds().Size()
	boardWidth := int(g.gameZoneSize.Width) * boardBlock.Y
	text.Draw(screen, "Счёт", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2, color.White)
	text.Draw(screen, fmt.Sprintf("%08d", g.score), g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+8, color.White)

	if g.state == GameStateGameOver {
		dy := 122
		text.Draw(screen, "GAME OVER", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy, color.White)
		text.Draw(screen, "Нажмите", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy+10, color.White)
		text.Draw(screen, "большую", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy+20, color.White)
		text.Draw(screen, "кнопку", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy+30, color.White)
	}

	if g.attractMode {
		dy := 148
		text.Draw(screen, "Нажмите", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy, color.White)
		text.Draw(screen, "большую", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy+10, color.White)
		text.Draw(screen, "кнопку...", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy+20, color.White)
	}

	dy := 48
	text.Draw(screen, "Следующая", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy, color.White)
	text.Draw(screen, "фигура", g.txtFont, boardWidth+gameZonePos.X*2, gameZonePos.Y*2+dy+10, color.White)
}

func (g *Game) updateScore(lines int) {
	perLineScore := 100
	g.score += lines * perLineScore
	if lines > 1 {
		bonus := perLineScore / 2
		for i := 0; i < int(lines); i++ {
			g.score += bonus
			bonus *= 2
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	gameZonePos := &Position{X: 16, Y: 16}

	g.drawText(screen, gameZonePos)

	gameZone := g.gameZone
	for y, row := range gameZone {
		for x, cellImage := range row {
			if cellImage == nil {
				cellImage = g.bgBlockImage
			}

			s := cellImage.Bounds().Size()
			screenPos := &Position{
				X: gameZonePos.X + x*s.X,
				Y: gameZonePos.Y + y*s.Y,
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(screenPos.X), float64(screenPos.Y))
			screen.DrawImage(cellImage, op)
		}
	}

	if g.currentPiece != nil {
		g.currentPiece.Draw(screen, gameZonePos, g.piecePosition)
	}

	if g.nextPiece != nil {
		nextPos := &Position{X: int(math.Round(ScreenWidth * 0.7)), Y: int(math.Round(ScreenHeight * .3))}
		g.nextPiece.Draw(screen, nextPos, &Position{})
	}

	// Рисуем маленькие кнопки (треугольник вершиной вниз)
	for _, btn := range g.smallButtons {
		btn.Draw(screen)
	}

	// Рисуем большую кнопку
	g.largeButton.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// SetupButtons создает и размещает кнопки согласно требованиям
func (g *Game) SetupButtons() {
	// Желтый цвет для кнопок
	yellow := color.RGBA{255, 255, 0, 255}

	// Параметры для маленьких кнопок
	smallRadius := float64(ScreenWidth * 0.08)
	margin := float64(ScreenWidth * 0.04) // Отступ от краев экрана

	// Вычисляем позиции для треугольника (вершиной вниз)
	baseY := float64(g.screenHeight) - margin*2 - smallRadius*2
	leftX := margin + smallRadius
	rightX := margin*2 + leftX + smallRadius*2
	topX := leftX + smallRadius*1.5
	topY := baseY - smallRadius*1.5
	bottomX := topX
	bottomY := baseY + smallRadius*1.5

	// Создаем маленькие кнопки
	g.smallButtons = []*Button{
		NewButton(leftX, baseY, smallRadius, yellow),     // Левая нижняя
		NewButton(topX, topY, smallRadius, yellow),       // Верхняя вершина
		NewButton(rightX, baseY, smallRadius, yellow),    // Правая нижняя
		NewButton(bottomX, bottomY, smallRadius, yellow), // Нижняя вершина
	}

	// Создаем большую кнопку справа
	largeRadius := smallRadius * 1.5
	largeX := float64(g.screenWidth) - margin - largeRadius
	largeY := float64(g.screenHeight) - margin - largeRadius

	g.largeButton = NewButton(largeX, largeY, largeRadius, yellow)
}

func NewGame() *Game {
	ebiten.SetTPS(18)

	game := &Game{
		txtFont:          NewFont(),
		inputAttractMode: NewAttractModeInput(),
		inputKeyboard:    &KeyboardInput{},
		dropTicks:        4,
		pieces:           allPieces,
		gameZoneSize:     Size{Width: 14, Height: 24},
		bgBlockImage:     createImage(imgBlockBG),
	}

	game.Start()

	return game
}
