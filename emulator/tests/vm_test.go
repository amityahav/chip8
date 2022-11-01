package tests

import (
	"chip8/emulator"
	"testing"
)

func Test(t *testing.T) {

	chip8 := emulator.VM{}
	t.Run("handle0", func(t *testing.T) {
		rom := []byte{0x00, 0xE0, 0x00, 0xEE}
		chip8.Init(rom[:1])
		//
	})

}
