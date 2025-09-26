package gtris

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const keyDown = ebiten.KeyDown

const SleepTouchMilliSec = 100

var inputKeys = []ebiten.Key{keyDown, ebiten.KeyLeft, ebiten.KeyRight, ebiten.KeyUp}

type Input interface {
	Read() *ebiten.Key
	IsSpacePressed() bool
}

type KeyboardInput struct{}

func (*KeyboardInput) IsSpacePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace)
}

func (*KeyboardInput) Read() *ebiten.Key {
	for _, key := range inputKeys {
		if key == ebiten.KeyUp {
			if inpututil.IsKeyJustPressed(key) {
				return &key
			}
		} else if ebiten.IsKeyPressed(key) {
			return &key
		}
	}

	return nil
}

type AttractModeInput struct {
	keyPressed chan ebiten.Key
}

func (input *AttractModeInput) IsSpacePressed() bool {
	// if there's a key available we just say it's space (this is only called to start the game)
	hasKey := input.Read() != nil
	return hasKey
}

func (input *AttractModeInput) Read() *ebiten.Key {
	select {
	case key := <-input.keyPressed:
		return &key
	default:
		return nil
	}
}

func NewAttractModeInput() *AttractModeInput {
	input := &AttractModeInput{
		keyPressed: make(chan ebiten.Key),
	}
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				key := keyDown
				if rand.Float32() < 0.5 {
					key = inputKeys[rand.Intn(len(inputKeys))]
				}
				input.keyPressed <- key
			}
		}
	}()

	return input
}

type TouchInput struct {
	smallButtons []*Button
	largeButton  *Button
}

func (t *TouchInput) SetupButtons(screenWidth, screenHeight int) {
	// Желтый цвет для кнопок
	yellow := color.RGBA{255, 255, 0, 255}

	// Параметры для маленьких кнопок
	smallRadius := float64(ScreenWidth * 0.1)
	margin := float64(ScreenWidth * 0.04) // Отступ от краев экрана

	// Вычисляем позиции для треугольника (вершиной вниз)
	baseY := float64(screenHeight) - margin*4 - smallRadius*2
	leftX := margin + smallRadius
	rightX := margin*4 + leftX + smallRadius*2
	bottomX := leftX + smallRadius*1.8
	bottomY := baseY + smallRadius*2.5

	// Создаем маленькие кнопки
	t.smallButtons = []*Button{
		NewButton(leftX, baseY, smallRadius, yellow, ebiten.KeyLeft),     // Левая нижняя
		NewButton(rightX, baseY, smallRadius, yellow, ebiten.KeyRight),   // Правая нижняя
		NewButton(bottomX, bottomY, smallRadius, yellow, ebiten.KeyDown), // Нижняя вершина
	}

	// Создаем большую кнопку справа
	largeRadius := smallRadius * 1.5
	largeX := float64(screenWidth) - margin - largeRadius
	largeY := float64(screenHeight) - margin - largeRadius

	t.largeButton = NewButton(largeX, largeY, largeRadius, yellow, ebiten.KeySpace)
}

func (t *TouchInput) Read() *ebiten.Key {
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
	for _, btn := range t.smallButtons {
		btn.Update(cursorX, cursorY, isMousePressed)
	}

	// Обновляем состояние большой кнопки
	t.largeButton.Update(cursorX, cursorY, isMousePressed)

	// В методе Update после обновления кнопок:
	for _, btn := range t.smallButtons {
		if btn.Pressed {
			if t.largeButton.PressedAgo() < SleepTouchMilliSec {
				return nil
			}
			t.largeButton.LastPressedTime = time.Now().UnixMilli()
			return &btn.key
		}
	}
	if t.largeButton.Pressed {
		if t.largeButton.PressedAgo() < SleepTouchMilliSec {
			return nil
		}
		t.largeButton.LastPressedTime = time.Now().UnixMilli()
		return &t.largeButton.key
	}
	return nil
}

func (t *TouchInput) IsSpacePressed() bool {
	return t.largeButton.Pressed
}

func (t *TouchInput) Draw(screen *ebiten.Image) {

	// Рисуем маленькие кнопки (треугольник вершиной вниз)
	for _, btn := range t.smallButtons {
		btn.Draw(screen)
	}

	// Рисуем большую кнопку
	t.largeButton.Draw(screen)
}
