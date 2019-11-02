package c8emu

import (
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
)

var (
	fonts = []uint8{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}
	whiteColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	blackColor = color.RGBA{R: 0, G: 0, B: 0, A: 255}
)

const (
	DisplayWidth  = 64
	DisplayHeight = 32
)

type c8Emu struct {
	memory      [0xFFF]uint8
	display     [DisplayHeight][DisplayWidth]uint8
	key         [16]uint8
	v           [16]uint8
	pc          uint16
	i           uint16
	sp          uint8
	stack       [16]uint16
	dt          uint8
	st          uint8
	shouldDraw  bool
	shouldSound bool
}

func New() *c8Emu {
	emu := c8Emu{
		shouldDraw:  true,
		shouldSound: true,
		pc:          0x200,
	}

	for i := 0; i < len(fonts); i++ {
		emu.memory[i] = fonts[i]
	}

	return &emu
}

func (emu *c8Emu) ShouldDraw() bool {
	sd := emu.shouldDraw
	emu.shouldDraw = false
	return sd
}

func (emu *c8Emu) ShouldSound() bool {
	ss := emu.shouldSound
	emu.shouldSound = false
	return ss
}

func (emu *c8Emu) SetKey(i uint8, active bool) {
	if active {
		emu.key[i] = 1
	} else {
		emu.key[i] = 0
	}
}

func (emu *c8Emu) GetDisplay() *[DisplayHeight][DisplayWidth]uint8 {
	return &emu.display
}

func (emu *c8Emu) GetDisplayImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, DisplayWidth, DisplayHeight))

	for y := 0; y < DisplayHeight; y++ {
		for x := 0; x < DisplayWidth; x++ {
			if emu.display[y][x] != 0 {
				img.SetRGBA(x, y, whiteColor)
			} else {
				img.SetRGBA(x, y, blackColor)
			}
		}
	}

	return img
}

func (emu *c8Emu) LoadFromFile(name string) {
	file, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(data); i++ {
		emu.memory[i+512] = data[i]
	}
}

func (emu *c8Emu) Cycle() {
	op := (uint16(emu.memory[emu.pc]) << 8) | uint16(emu.memory[emu.pc+1])

	switch op & 0xF000 {

	case 0x0000:
		switch op & 0x000F {

		/*
			00E0 - CLS
			Clear the display.
		*/
		case 0x0000:
			for y := 0; y < DisplayHeight; y++ {
				for x := 0; x < DisplayWidth; x++ {
					emu.display[y][x] = 0x0
				}
			}

			emu.shouldDraw = true
			emu.pc = emu.pc + 2

		/*
			00EE - RET
			Return from a subroutine.
		*/
		case 0x000E:
			emu.sp = emu.sp - 1
			emu.pc = emu.stack[emu.sp]
			emu.pc = emu.pc + 2

		default:
			log.Fatalf("c8emu: invalid op %X\n", op)

		}

	/*
		1nnn - JP addr
		Jump to location nnn.
	*/
	case 0x1000:
		emu.pc = op & 0x0FFF

	/*
		2nnn - CALL addr
		Call subroutine at nnn.
	*/
	case 0x2000:
		emu.stack[emu.sp] = emu.pc
		emu.sp = emu.sp + 1
		emu.pc = op & 0x0FFF

	/*
		3xkk - SE Vx, byte
		Skip next instruction if Vx = kk.
	*/
	case 0x3000:
		if uint16(emu.v[(op&0x0F00)>>8]) == op&0x00FF {
			emu.pc = emu.pc + 4
		} else {
			emu.pc = emu.pc + 2
		}

	/*
		4xkk - SNE Vx, byte
		Skip next instruction if Vx != kk.
	*/
	case 0x4000:
		if uint16(emu.v[(op&0x0F00)>>8]) != op&0x00FF {
			emu.pc = emu.pc + 4
		} else {
			emu.pc = emu.pc + 2
		}

	/*
		5xy0 - SE Vx, Vy
		Skip next instruction if Vx = Vy.
	*/
	case 0x5000:
		if emu.v[(op&0x0F00)>>8] == emu.v[(op&0x00F0)>>4] {
			emu.pc = emu.pc + 4
		} else {
			emu.pc = emu.pc + 2
		}

	/*
		6xkk - LD Vx, byte
		Set Vx = kk.
	*/
	case 0x6000:
		emu.v[(op&0x0F00)>>8] = uint8(op & 0x00FF)
		emu.pc = emu.pc + 2

	/*
		7xkk - ADD Vx, byte
		Set Vx = Vx + kk.
	*/
	case 0x7000:
		emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] + uint8(op&0x00FF)
		emu.pc = emu.pc + 2

	case 0x8000:
		switch op & 0x000F {

		/*
			8xy0 - LD Vx, Vy
			Set Vx = Vy.
		*/
		case 0x0000:
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x00F0)>>4]
			emu.pc = emu.pc + 2

		/*
			8xy1 - OR Vx, Vy
			Set Vx = Vx OR Vy.
		*/
		case 0x0001:
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] | emu.v[(op&0x00F0)>>4]
			emu.pc = emu.pc + 2

		/*
			8xy2 - AND Vx, Vy
			Set Vx = Vx AND Vy.
		*/
		case 0x0002:
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] & emu.v[(op&0x00F0)>>4]
			emu.pc = emu.pc + 2

		/*
			8xy3 - XOR Vx, Vy
			Set Vx = Vx XOR Vy.
		*/
		case 0x0003:
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] ^ emu.v[(op&0x00F0)>>4]
			emu.pc = emu.pc + 2

		/*
			8xy4 - ADD Vx, Vy
			Set Vx = Vx + Vy, set VF = carry.
		*/
		case 0x0004:
			if emu.v[(op&0x00F0)>>4] > 0xFF-emu.v[(op&0x0F00)>>8] {
				emu.v[0xF] = 1
			} else {
				emu.v[0xF] = 0
			}
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] + emu.v[(op&0x00F0)>>4]
			emu.pc = emu.pc + 2

		/*
			8xy5 - SUB Vx, Vy
			Set Vx = Vx - Vy, set VF = NOT borrow.
		*/
		case 0x0005:
			if emu.v[(op&0x00F0)>>4] > emu.v[(op&0x0F00)>>8] {
				emu.v[0xF] = 0
			} else {
				emu.v[0xF] = 1
			}
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] - emu.v[(op&0x00F0)>>4]
			emu.pc = emu.pc + 2

		/*
			8xy6 - SHR Vx {, Vy}
			Set Vx = Vx SHR 1.
		*/
		case 0x0006:
			emu.v[0xF] = emu.v[(op&0x0F00)>>8] & 0x1
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] >> 1
			emu.pc = emu.pc + 2

		/*
			8xy7 - SUBN Vx, Vy
			Set Vx = Vy - Vx, set VF = NOT borrow.
		*/
		case 0x0007:
			if emu.v[(op&0x0F00)>>8] > emu.v[(op&0x00F0)>>4] {
				emu.v[0xF] = 0
			} else {
				emu.v[0xF] = 1
			}
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x00F0)>>4] - emu.v[(op&0x0F00)>>8]
			emu.pc = emu.pc + 2

		/*
			8xyE - SHL Vx {, Vy}
			Set Vx = Vx SHL 1.
		*/
		case 0x000E:
			emu.v[0xF] = emu.v[(op&0x0F00)>>8] >> 7
			emu.v[(op&0x0F00)>>8] = emu.v[(op&0x0F00)>>8] << 1
			emu.pc = emu.pc + 2

		default:
			log.Fatalf("c8emu: invalid op %X\n", op)

		}

	/*
		9xy0 - SNE Vx, Vy
		Skip next instruction if Vx != Vy.
	*/
	case 0x9000:
		if emu.v[(op&0x0F00)>>8] != emu.v[(op&0x00F0)>>4] {
			emu.pc = emu.pc + 4
		} else {
			emu.pc = emu.pc + 2
		}

	/*
		Annn - LD I, addr
		Set I = nnn.
	*/
	case 0xA000:
		emu.i = op & 0x0FFF
		emu.pc = emu.pc + 2

	/*
		Bnnn - JP V0, addr
		Jump to location nnn + V0.
	*/
	case 0xB000:
		emu.pc = (op & 0x0FFF) + uint16(emu.v[0x0])

	/*
		Cxkk - RND Vx, byte
		Set Vx = random byte AND kk.
	*/
	case 0xC000:
		emu.v[(op&0x0F00)>>8] = uint8(rand.Intn(256)) & uint8(op&0x00FF)
		emu.pc = emu.pc + 2

	/*
		Dxyn - DRW Vx, Vy, nibble
		Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
	*/
	case 0xD000:
		rx := emu.v[(op&0x0F00)>>8]
		ry := emu.v[(op&0x00F0)>>4]
		h := op & 0x000F

		emu.v[0xF] = 0

		var yl uint16 = 0
		var xl uint16 = 0

		for yl = 0; yl < h; yl++ {
			row := emu.memory[emu.i+yl]

			for xl = 0; xl < 8; xl++ {
				if (row & (0x80 >> xl)) != 0 {
					x := rx + uint8(xl)
					if x >= DisplayWidth {
						x -= DisplayWidth
					}

					y := ry + uint8(yl)
					if y >= DisplayHeight {
						y -= DisplayHeight
					}

					if emu.display[y][x] == 1 {
						emu.v[0xF] = 1
					}

					emu.display[y][x] ^= 1
				}
			}
		}

		emu.shouldDraw = true
		emu.pc = emu.pc + 2

	case 0xE000:
		switch op & 0x00FF {

		/*
			Ex9E - SKP Vx
			Skip next instruction if key with the value of Vx is pressed.
		*/
		case 0x009E:
			if emu.key[emu.v[(op&0x0F00)>>8]] == 1 {
				emu.pc = emu.pc + 4
			} else {
				emu.pc = emu.pc + 2
			}

		/*
			ExA1 - SKNP Vx
			Skip next instruction if key with the value of Vx is not pressed.
		*/
		case 0x00A1:
			if emu.key[emu.v[(op&0x0F00)>>8]] == 0 {
				emu.pc = emu.pc + 4
			} else {
				emu.pc = emu.pc + 2
			}

		default:
			log.Fatalf("c8emu: invalid op %X\n", op)

		}

	case 0xF000:
		switch op & 0x00FF {

		/*
			Fx07 - LD Vx, DT
			Set Vx = delay timer value.
		*/
		case 0x0007:
			emu.v[(op&0x0F00)>>8] = emu.dt
			emu.pc = emu.pc + 2

		/*
			Fx0A - LD Vx, K
			Wait for a key press, store the value of the key in Vx.
		*/
		case 0x000A:
			pressed := false
			for i := 0; i < len(emu.key); i++ {
				if emu.key[i] != 0 {
					emu.v[(op&0x0F00)>>8] = uint8(i)
					pressed = true
				}
			}
			if !pressed {
				return
			}
			emu.pc = emu.pc + 2

		/*
			Fx15 - LD DT, Vx
			Set delay timer = Vx.
		*/
		case 0x0015:
			emu.dt = emu.v[(op&0x0F00)>>8]
			emu.pc = emu.pc + 2

		/*
			Fx18 - LD ST, Vx
			Set sound timer = Vx.
		*/
		case 0x0018:
			emu.st = emu.v[(op&0x0F00)>>8]
			emu.pc = emu.pc + 2

		/*
			Fx1E - ADD I, Vx
			Set I = I + Vx.
		*/
		case 0x001E:
			if emu.i+uint16(emu.v[(op&0x0F00)>>8]) > 0xFFF {
				emu.v[0xF] = 1
			} else {
				emu.v[0xF] = 0
			}
			emu.i = emu.i + uint16(emu.v[(op&0x0F00)>>8])
			emu.pc = emu.pc + 2

		/*
			Fx29 - LD F, Vx
			Set I = location of sprite for digit Vx.
		*/
		case 0x0029:
			emu.i = uint16(emu.v[(op&0x0F00)>>8]) * 0x5
			emu.pc = emu.pc + 2

		/*
			Fx33 - LD B, Vx
			Store BCD representation of Vx in memory locations I, I+1, and I+2.
		*/
		case 0x0033:
			emu.memory[emu.i] = emu.v[(op&0x0F00)>>8] / 100
			emu.memory[emu.i+1] = (emu.v[(op&0x0F00)>>8] / 10) % 10
			emu.memory[emu.i+2] = (emu.v[(op&0x0F00)>>8] % 100) / 10
			emu.pc = emu.pc + 2

		/*
			Fx55 - LD [I], Vx
			Store registers V0 through Vx in memory starting at location I.
		*/
		case 0x0055:
			for i := 0; i < int((op&0x0F00)>>8)+1; i++ {
				emu.memory[uint16(i)+emu.i] = emu.v[i]
			}
			emu.i = ((op & 0x0F00) >> 8) + 1
			emu.pc = emu.pc + 2

		/*
			Fx65 - LD Vx, [I]
			Read registers V0 through Vx from memory starting at location I.
		*/
		case 0x0065:
			for i := 0; i < int((op&0x0F00)>>8)+1; i++ {
				emu.v[i] = emu.memory[emu.i+uint16(i)]
			}
			emu.i = ((op & 0x0F00) >> 8) + 1
			emu.pc = emu.pc + 2

		default:
			log.Fatalf("c8emu: invalid op %X\n", op)

		}

	default:
		log.Fatalf("c8emu: invalid op %X\n", op)

	}

	if emu.dt > 0 {
		emu.dt--
	}

	if emu.st > 0 {
		if emu.st == 1 {
			emu.shouldSound = true
		}
		emu.st--
	}
}
