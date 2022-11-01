package main

import (
	"chip8/emulator"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/veandco/go-sdl2/sdl"
	"io/ioutil"
	"log"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chip8",
		Short: "Chip8 Emulator written in Go",
	}

	play := &cobra.Command{
		Use:   "play",
		Short: "specify path of ROM you want to play",
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

			window, err := sdl.CreateWindow(fmt.Sprintf("Chip8 Emulator - %s", fileName), sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
				sdl.WINDOWPOS_UNDEFINED, emulator.Width, emulator.Height)
			if err != nil {
				log.Fatalf("main: failed to create sdl window err: %s", err.Error())
			}
			defer window.Destroy()

			surface, err := window.GetSurface()
			if err != nil {
				log.Fatalf("main: failed to get surface err: %s", err.Error())
			}

			surface.FillRect(nil, 0)

			rect := sdl.Rect{0, 0, 200, 200}
			surface.FillRect(&rect, 0xffff0000)
			window.UpdateSurface()

			running := true
			for running {
				for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
					switch event.(type) {
					case *sdl.QuitEvent:
						println("Quit")
						running = false
						break
					}
				}
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
