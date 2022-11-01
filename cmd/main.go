package main

import (
	"chip8/emulator"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/veandco/go-sdl2/sdl"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chip8",
		Short: "Chip8 Emulator written in Go",
	}

	play := &cobra.Command{
		Use:   "play",
		Short: "specify name of the ROM you want to play",
		Run: func(cmd *cobra.Command, args []string) {
			fileName := args[0]
			file := romToBytes(fmt.Sprintf("./roms/%s", fileName))

			// Chip8 Init
			chip8 := emulator.VM{}
			chip8.Init(file)

			// SDL Init
			if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
				log.Fatalf("main: failed to init sdl err: %s", err.Error())
			}
			defer sdl.Quit()

			window, err := sdl.CreateWindow(fmt.Sprintf("Chip8 Emulator - %s", fileName), sdl.WINDOWPOS_UNDEFINED,
				sdl.WINDOWPOS_UNDEFINED, emulator.Width*emulator.Factor, emulator.Height*emulator.Factor, sdl.WINDOW_SHOWN)
			if err != nil {
				log.Fatalf("main: failed to create sdl window err: %s", err.Error())
			}
			defer window.Destroy()

			canvas, err := sdl.CreateRenderer(window, -1, 0)
			if err != nil {
				log.Fatalf("main: failed to create renderer err: %s", err.Error())
			}
			defer canvas.Destroy()

			// Main Loop
			for {
				// Execute next Opcode
				chip8.DecAndExec()

				// Draw only when needed
				if chip8.Draw() {
					for i := 0; i < len(chip8.Screen); i++ {
						for j := 0; j < len(chip8.Screen[i]); j++ {
							if chip8.Screen[i][j] != 0 {
								canvas.SetDrawColor(255, 255, 255, 0)
							} else {
								canvas.SetDrawColor(0, 0, 0, 0)
							}

							canvas.FillRect(&sdl.Rect{
								X: int32(i) * emulator.Factor,
								Y: int32(i) * emulator.Factor,
								W: emulator.Factor,
								H: emulator.Factor,
							})
						}
					}
					canvas.Present()
				}

				for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
					switch et := event.(type) {
					case *sdl.QuitEvent:
						os.Exit(0)
					case *sdl.KeyboardEvent:
						chip8.Key(et.Keysym.Sym, et.Type)
					}
				}

				sdl.Delay(60 / 1000) // 60 Hz
			}
		},
	}

	rootCmd.AddCommand(play)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln("main: failed to start emulator")
	}
}

func romToBytes(path string) []byte {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("romToBytes: failed to read ROM %s err: %s", path, err.Error())
	}

	return file
}
