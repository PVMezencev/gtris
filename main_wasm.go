//go:build js && wasm
// +build js,wasm

package main

import (
	"log"
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/luisparravicini/gtris/gtris"
)

func main() {
	game := gtris.NewGame()
	game.SetScreenWidth(gtris.ScreenWidth)
	game.SetScreenHeight(gtris.ScreenHeight)

	js.Global().Set("getScore", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		// Предполагается, что у game есть метод для получения счета
		return js.ValueOf(game.Score())
	}))

	ebiten.SetWindowSize(gtris.ScreenWidth, gtris.ScreenHeight)
	ebiten.SetWindowTitle("gtris")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
