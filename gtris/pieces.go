package gtris

import (
	_ "embed"
	_ "image/png"
)

//go:embed images/block-a.png
var imgBlockA []byte

//go:embed images/block-b.png
var imgBlockB []byte

//go:embed images/block-c.png
var imgBlockC []byte

//go:embed images/block-d.png
var imgBlockD []byte

//go:embed images/block-e.png
var imgBlockE []byte

//go:embed images/block-f.png
var imgBlockF []byte

//go:embed images/block-g.png
var imgBlockG []byte

//go:embed images/block-bg.png
var imgBlockBG []byte

var allPieces = []*Piece{
	NewPiece([][]int{
		{1, 1, 1, 1},
		{0, 0, 0, 0},
	}, imgBlockA, FTypeA),
	NewPiece([][]int{
		{1, 0, 0, 0},
		{1, 1, 1, 1},
	}, imgBlockB, FTypeB),
	NewPiece([][]int{
		{0, 0, 0, 1},
		{1, 1, 1, 1},
	}, imgBlockC, FTypeC),
	NewPiece([][]int{
		{1, 1, 0, 0},
		{1, 1, 0, 0},
	}, imgBlockD, FTypeD),
	NewPiece([][]int{
		{0, 1, 1, 0},
		{1, 1, 0, 0},
	}, imgBlockE, FTypeE),
	NewPiece([][]int{
		{0, 1, 0, 0},
		{1, 1, 1, 0},
	}, imgBlockF, FTypeF),
	NewPiece([][]int{
		{1, 1, 0, 0},
		{0, 1, 1, 0},
	}, imgBlockG, FTypeG),
}
