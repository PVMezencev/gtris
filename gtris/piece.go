package gtris

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const pieceBlockMarker = 1

const (
	FTypeA = "A"
	FTypeB = "B"
	FTypeC = "C"
	FTypeD = "D"
	FTypeE = "E"
	FTypeF = "F"
	FTypeG = "G"
)

type Piece struct {
	Blocks [][]int
	Image  *ebiten.Image
	FType  string
}

func NewPiece(blocks [][]int, imgData []byte, fType string) *Piece {
	return &Piece{
		Blocks: blocks,
		Image:  createImage(imgData),
		FType:  fType,
	}
}

func (p *Piece) Rotate() {
	if p.FType == FTypeD {
		// Квадрат не поворачиваем.
		return
	}
	new_matrix := [][]int{}
	for i := 0; i < len(p.Blocks[0]); i++ {
		new_matrix = append(new_matrix, []int{})
	}
	for i := len(p.Blocks) - 1; i >= 0; i-- {
		for j, dat := range p.Blocks[i] {
			new_matrix[j] = append(new_matrix[j], dat)
		}
	}
	p.Blocks = new_matrix
}

func (p *Piece) Draw(screen *ebiten.Image, gameZonePos *Position, piecePos *Position) {
	size := p.Image.Bounds().Size()
	w, h := size.X, size.Y
	for dy, row := range p.Blocks {
		for dx, value := range row {
			if value == pieceBlockMarker {
				screenPos := &Position{
					X: gameZonePos.X + (piecePos.X+dx)*w,
					Y: gameZonePos.Y + (piecePos.Y+dy)*h,
				}
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(screenPos.X), float64(screenPos.Y))
				screen.DrawImage(p.Image, op)
			}
		}
	}
}
