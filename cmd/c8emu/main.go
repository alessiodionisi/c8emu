package main

// typedef unsigned char Uint8;
// void sineWave(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"github.com/adnsio/c8emu"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math"
	"reflect"
	"time"
	"unsafe"
)

const (
	scale = 10
)

var (
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
)

//export sineWave
func sineWave(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]C.Uint8)(unsafe.Pointer(&hdr))

	var phase float64
	for i := 0; i < n; i += 2 {
		phase += 2 * math.Pi * 440 / 44100
		sample := C.Uint8((math.Sin(phase) + 0.999999) * 128)
		buf[i] = sample
		buf[i+1] = sample
	}
}

func init() {
	var err error

	if err = sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO); err != nil {
		log.Fatal(err)
	}

	window, err = sdl.CreateWindow("c8emu", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, c8emu.DisplayWidth*scale, c8emu.DisplayHeight*scale, sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatal(err)
	}

	renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatal(err)
	}

	texture, err = renderer.CreateTexture(uint32(sdl.PIXELFORMAT_RGBA32), sdl.TEXTUREACCESS_STREAMING, c8emu.DisplayWidth, c8emu.DisplayHeight)
	if err != nil {
		log.Fatal(err)
	}

	audioSpec := &sdl.AudioSpec{
		Freq:     44100,
		Format:   sdl.AUDIO_S16SYS,
		Channels: 1,
		Samples:  2048,
		Callback: sdl.AudioCallback(C.sineWave),
	}

	if err := sdl.OpenAudio(audioSpec, nil); err != nil {
		log.Fatal(err)
	}
}

func playSound() {
	sdl.PauseAudio(false)

	go func() {
		t := time.NewTimer(time.Second / time.Duration(10))
		select {
		case <-t.C:
			sdl.PauseAudio(true)
		}
	}()
}

func main() {
	defer sdl.Quit()
	defer window.Destroy()
	defer renderer.Destroy()
	defer texture.Destroy()
	defer sdl.CloseAudio()

	playSound()

	emu := c8emu.New()
	emu.LoadFromFile("roms/PONG2")

	ticker := time.NewTicker(time.Second / time.Duration(60))
	running := true

	for range ticker.C {
		if !running {
			break
		}

		emu.Cycle()

		if emu.ShouldSound() {
			playSound()
		}

		if emu.ShouldDraw() {
			if err := renderer.SetDrawColor(0, 0, 0, 0); err != nil {
				log.Fatal(err)
			}

			if err := renderer.Clear(); err != nil {
				log.Fatal(err)
			}

			img := emu.GetDisplayImage()

			if err := texture.Update(nil, img.Pix, img.Stride); err != nil {
				log.Fatal(err)
			}

			if err := renderer.Copy(texture, nil, nil); err != nil {
				log.Fatal(err)
			}

			renderer.Present()
		}

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {

			case *sdl.QuitEvent:
				running = false
				break

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYUP {
					switch e.Keysym.Sym {

					case sdl.K_1:
						emu.SetKey(0x1, false)
					case sdl.K_2:
						emu.SetKey(0x2, false)
					case sdl.K_3:
						emu.SetKey(0x3, false)
					case sdl.K_4:
						emu.SetKey(0xC, false)
					case sdl.K_q:
						emu.SetKey(0x4, false)
					case sdl.K_w:
						emu.SetKey(0x5, false)
					case sdl.K_e:
						emu.SetKey(0x6, false)
					case sdl.K_r:
						emu.SetKey(0xD, false)
					case sdl.K_a:
						emu.SetKey(0x7, false)
					case sdl.K_s:
						emu.SetKey(0x8, false)
					case sdl.K_d:
						emu.SetKey(0x9, false)
					case sdl.K_f:
						emu.SetKey(0xE, false)
					case sdl.K_z:
						emu.SetKey(0xA, false)
					case sdl.K_x:
						emu.SetKey(0x0, false)
					case sdl.K_c:
						emu.SetKey(0xB, false)
					case sdl.K_v:
						emu.SetKey(0xF, false)
					}

				} else if e.Type == sdl.KEYDOWN {
					switch e.Keysym.Sym {

					case sdl.K_1:
						emu.SetKey(0x1, true)
					case sdl.K_2:
						emu.SetKey(0x2, true)
					case sdl.K_3:
						emu.SetKey(0x3, true)
					case sdl.K_4:
						emu.SetKey(0xC, true)
					case sdl.K_q:
						emu.SetKey(0x4, true)
					case sdl.K_w:
						emu.SetKey(0x5, true)
					case sdl.K_e:
						emu.SetKey(0x6, true)
					case sdl.K_r:
						emu.SetKey(0xD, true)
					case sdl.K_a:
						emu.SetKey(0x7, true)
					case sdl.K_s:
						emu.SetKey(0x8, true)
					case sdl.K_d:
						emu.SetKey(0x9, true)
					case sdl.K_f:
						emu.SetKey(0xE, true)
					case sdl.K_z:
						emu.SetKey(0xA, true)
					case sdl.K_x:
						emu.SetKey(0x0, true)
					case sdl.K_c:
						emu.SetKey(0xB, true)
					case sdl.K_v:
						emu.SetKey(0xF, true)
					}
				}
			}
		}
	}
}
