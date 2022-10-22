package src

import "log"

type VM struct {
	memory          [memorySize]byte
	screen          [32][8]byte // 32x64 bitmap
	stack           [16]uint16
	v               [16]byte
	dt, st, sp      byte
	pc, opcode, I   uint16
	opcodesHandlers map[uint16]any
}

func (vm *VM) init(bytes []byte) {
	vm.loadROM(bytes)
	vm.pc = startAddress
	vm.opcodesHandlers = map[uint16]any{
		0x0000: vm.handle0,
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

func (vm *VM) decAndExec() {
	vm.opcode = uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc])
	vm.opcodesHandlers[vm.opcode&0xF000]()
}

// Handlers
func (vm *VM) handle0() {
	switch vm.opcode & 0x00FF {
	case 0x0E0: // Clear the display

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
