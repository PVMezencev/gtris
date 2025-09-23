package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/luisparravicini/gtris/gtris"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	game := gtris.NewGame()
	game.SetScreenWidth(gtris.ScreenWidth)
	game.SetScreenHeight(gtris.ScreenHeight)
	game.SetupButtons()

	ebiten.SetWindowSize(gtris.ScreenWidth, gtris.ScreenHeight)
	ebiten.SetWindowTitle("gtris")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
