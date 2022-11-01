package emulator

import (
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math/rand"
	"time"
)

type keyboard struct {
	keys    [16]byte
	keysMap map[sdl.Keycode]byte
}

func (k *keyboard) init() {
	k.keysMap = map[sdl.Keycode]byte{
		sdl.K_1: 0x01,
		sdl.K_2: 0x02,
		sdl.K_3: 0x03,
		sdl.K_4: 0x0C,
		sdl.K_q: 0x04,
		sdl.K_w: 0x05,
		sdl.K_e: 0x06,
		sdl.K_r: 0x0D,
		sdl.K_a: 0x07,
		sdl.K_s: 0x08,
		sdl.K_d: 0x09,
		sdl.K_f: 0x0E,
		sdl.K_z: 0x0A,
		sdl.K_x: 0x00,
		sdl.K_c: 0x08,
		sdl.K_v: 0x0F,
	}
}

type VM struct {
	memory          [memorySize]byte
	Screen          [32][8]byte // 32x64 bitmap
	stack           [16]uint16
	v               [16]byte
	keyboard        keyboard
	dt, st, sp      byte
	pc, opcode, i   uint16
	opcodesHandlers map[uint16]func()
	shouldDraw      bool
}

func (vm *VM) Init(bytes []byte) {
	vm.loadROM(bytes)
	vm.initFonts()
	vm.keyboard.init()
	vm.pc = startAddress
	vm.shouldDraw = true
	vm.opcodesHandlers = map[uint16]func(){
		0x0000: vm.handle0,
		0x1000: vm.handle1,
		0x2000: vm.handle2,
		0x3000: vm.handle3,
		0x4000: vm.handle4,
		0x5000: vm.handle5,
		0x6000: vm.handle6,
		0x7000: vm.handle7,
		0x8000: vm.handle8,
		0x9000: vm.handle9,
		0xA000: vm.handleA,
		0xB000: vm.handleB,
		0xC000: vm.handleC,
		0xD000: vm.handleD,
		0xE000: vm.handleE,
		0xF000: vm.handleF,
	}
}

func (vm *VM) loadROM(bytes []byte) {
	for i, b := range bytes {
		vm.memory[startAddress+i] = b
	}
}

func (vm *VM) initFonts() {
	for i, b := range fonts {
		vm.memory[i] = b
	}
}

func (vm *VM) Key(key sdl.Keycode, t sdl.EventType) {
	k := vm.keyboard.keysMap[key]

	if t == sdl.KEYUP {
		vm.keyboard.keys[k] = 0
	} else if t == sdl.KEYDOWN {
		vm.keyboard.keys[k] = 1
	}
}

// inBounds checks if idx is within v bounds and panics if not
func (vm *VM) inBounds(name string, idx uint16) {
	if idx < 0 || int(idx) >= len(vm.v) {
		log.Fatalf("handle3: %s is out of bounds", name)
	}
}

func (vm *VM) Draw() bool {
	sd := vm.shouldDraw
	vm.shouldDraw = false
	return sd
}

func (vm *VM) DecAndExec() {
	vm.opcode = uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc+1])
	vm.opcodesHandlers[vm.opcode&0xF000]()
}

// Handlers

// handle0 - CLS | RET
func (vm *VM) handle0() {
	switch vm.opcode & 0x00FF {
	case 0x0E0: // Clear the display
		for i := 0; i < len(vm.Screen); i++ {
			for j := 0; j < len(vm.Screen); j++ {
				vm.Screen[i][j] = 0
			}
		}
		vm.shouldDraw = true
		vm.pc += 2
	case 0x0EE: // Return from a subroutine
		if vm.sp < 1 {
			log.Fatal("handle0: stack pointer is < 1")
		}

		vm.sp--
		vm.pc = vm.stack[vm.sp]
	default:
		vm.pc += 2 // Ignoring opcode
	}
}

// handle1 - JP addr
func (vm *VM) handle1() {
	addr := vm.opcode & 0x0FFF
	vm.pc = addr
}

// handle2 - CALL addr
func (vm *VM) handle2() {
	addr := vm.opcode & 0x0FFF
	vm.stack[vm.sp] = vm.pc + 2
	vm.sp++
	vm.pc = addr
}

// handle3 - SE Vx, byte
func (vm *VM) handle3() {
	b := byte(vm.opcode & 0x0FF)
	x := (vm.opcode & 0x0F00) >> 8
	vm.inBounds("x", x)

	if vm.v[x] == b {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

// handle4 - SNE Vx, byte
func (vm *VM) handle4() {
	b := byte(vm.opcode & 0x0FF)
	x := (vm.opcode & 0x0F00) >> 8
	vm.inBounds("x", x)

	if vm.v[x] != b {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

// handle5 - SE Vx, Vy
func (vm *VM) handle5() {
	x := (vm.opcode & 0x0F00) >> 8
	y := (vm.opcode & 0x00F0) >> 4
	vm.inBounds("x", x)
	vm.inBounds("y", y)

	if vm.v[x] == vm.v[y] {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

// handle6 - LD Vx, byte
func (vm *VM) handle6() {
	b := byte(vm.opcode & 0x00FF)
	x := (vm.opcode & 0x0F00) >> 8
	vm.inBounds("x", x)

	vm.v[x] = b
	vm.pc += 2
}

// handle7 - ADD Vx, byte
func (vm *VM) handle7() {
	b := byte(vm.opcode & 0x00FF)
	x := (vm.opcode & 0x0F00) >> 8
	vm.inBounds("x", x)

	vm.v[x] += b
	vm.pc += 2
}

// handle8 - <LD | OR | AND | XOR | ADD | SUB | SHR | SUBN | SHL> Vx, Vy
func (vm *VM) handle8() {
	op := byte(vm.opcode & 0x000F)
	x := (vm.opcode & 0x0F00) >> 8
	y := (vm.opcode & 0x00F0) >> 4
	vm.inBounds("x", x)
	vm.inBounds("y", y)

	switch op {
	case 0x00: // LD
		vm.v[x] = vm.v[y]
	case 0x01: // OR
		vm.v[x] |= vm.v[y]
	case 0x02: // AND
		vm.v[x] &= vm.v[y]
	case 0x03: // XOR
		vm.v[x] ^= vm.v[y]
	case 0x04: // ADD
		vm.v[x] += vm.v[y]
	case 0x05: // SUB
		if vm.v[x] > vm.v[y] {
			vm.v[0x0F] = 1
		} else {
			vm.v[0x0F] = 0
		}
		vm.v[x] -= vm.v[y]
	case 0x06: // SHR
		if vm.v[x]&0x01 == 0x01 {
			vm.v[0x0F] = 1
		} else {
			vm.v[0x0F] = 0
		}
		vm.v[x] >>= 1
	case 0x07: // SUBN
		if vm.v[y] > vm.v[x] {
			vm.v[0x0F] = 1
		} else {
			vm.v[0x0F] = 0
		}
		vm.v[x] = vm.v[y] - vm.v[x]

	case 0x0E: // SHL
		mask := byte(0x01 << 7)
		if vm.v[x]&mask == 0x80 {
			vm.v[0x0F] = 1
		} else {
			vm.v[0x0F] = 0
		}
		vm.v[x] <<= 1
	}
	vm.pc += 2
}

// handle9 - SNE Vx, Vy
func (vm *VM) handle9() {
	x := (vm.opcode & 0x0F00) >> 8
	y := (vm.opcode & 0x00F0) >> 4
	vm.inBounds("x", x)
	vm.inBounds("y", y)

	if vm.v[x] != vm.v[y] {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

// handleA - LD I, addr
func (vm *VM) handleA() {
	addr := vm.opcode & 0x0FFF
	vm.i = addr
	vm.pc += 2
}

// handleB - JP V0, addr
func (vm *VM) handleB() {
	addr := vm.opcode & 0x0FFF
	vm.pc = uint16(vm.v[0]) + addr
}

// handleC - RND Vx, byte
func (vm *VM) handleC() {
	rand.Seed(time.Now().UnixNano())
	rn := byte(rand.Intn(256))
	x := (vm.opcode & 0x0F00) >> 8
	vm.inBounds("x", x)
	b := byte(vm.opcode & 0x00FF)

	vm.v[x] = rn & b
	vm.pc += 2
}

// handleD - DRW Vx, Vy, nibble
func (vm *VM) handleD() {
	//x := (vm.opcode & 0x0F00) >> 8
	//y := (vm.opcode & 0x00F0) >> 4
	//vm.inBounds("x", x)
	//vm.inBounds("y", y)
	//
	//n := vm.opcode & 0x000F
	//sprites := vm.memory[vm.i : vm.i+n]
	//
	////for i, b := range sprites {
	////	for j := 0; j < len(vm.Screen[0]); j++ {
	////		if x%8 != 0 { // Overlapping bytes
	////		vm.pc += 2
	////		}
	////	}
	////}
}

// handleE - <SKP | SKNP> Vx
func (vm *VM) handleE() {
	x := (vm.opcode & 0x0F00) >> 8
	vm.inBounds("x", x)
	op := vm.opcode & 0x00FF

	switch op {
	case 0x9E: // SKP
		if vm.keyboard.keys[vm.v[x]] == 1 {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
	case 0xA1: // SKNP
		if vm.keyboard.keys[vm.v[x]] == 0 {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
	}
}

// handleF
func (vm *VM) handleF() {
	x := (vm.opcode & 0x0F00) >> 8
	vm.inBounds("x", x)
	op := vm.opcode & 0x00FF

	switch op {
	case 0x07: // LD Vx, DT
		vm.v[x] = vm.dt
	case 0x0A: // LD Vx, K
		pressed := false

		for i, k := range vm.keyboard.keys {
			if k == 1 {
				vm.v[x] = byte(i)
				pressed = true
			}
		}

		if !pressed {
			return
		}
		vm.pc += 2
	}
}
