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

var (
	window         *sdl.Window
	renderer       *sdl.Renderer
	texture        *sdl.Texture
	interval       = uint32(1000 / 60)
	displayScale   = int32(10)
	maxCycles      = 12
	audioFrequency = 48000
	audioTone      = 440
)

//export sineWave
func sineWave(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]C.Uint8)(unsafe.Pointer(&hdr))

	var phase float64
	for i := 0; i < n; i += 2 {
		phase += 2 * math.Pi * float64(audioTone) / float64(audioFrequency)
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

	window, err = sdl.CreateWindow("c8emu", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, c8emu.DisplayWidth*displayScale, c8emu.DisplayHeight*displayScale, sdl.WINDOW_SHOWN)
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
		Freq:     int32(audioFrequency),
		Format:   sdl.AUDIO_U8,
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
		t := time.NewTimer(time.Second / time.Duration(60))
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
	emu.LoadFromFile("roms/BRIX")

	paused := false
	muted := false
	running := true

	for running {
		tick1 := sdl.GetTicks()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {

			case *sdl.QuitEvent:
				running = false
				break

			case *sdl.KeyboardEvent:
				active := e.Type == sdl.KEYUP

				switch e.Keysym.Sym {

				case sdl.K_1:
					emu.SetKey(0x1, active)
				case sdl.K_2:
					emu.SetKey(0x2, active)
				case sdl.K_3:
					emu.SetKey(0x3, active)
				case sdl.K_4:
					emu.SetKey(0xC, active)
				case sdl.K_q:
					emu.SetKey(0x4, active)
				case sdl.K_w:
					emu.SetKey(0x5, active)
				case sdl.K_e:
					emu.SetKey(0x6, active)
				case sdl.K_r:
					emu.SetKey(0xD, active)
				case sdl.K_a:
					emu.SetKey(0x7, active)
				case sdl.K_s:
					emu.SetKey(0x8, active)
				case sdl.K_d:
					emu.SetKey(0x9, active)
				case sdl.K_f:
					emu.SetKey(0xE, active)
				case sdl.K_z:
					emu.SetKey(0xA, active)
				case sdl.K_x:
					emu.SetKey(0x0, active)
				case sdl.K_c:
					emu.SetKey(0xB, active)
				case sdl.K_v:
					emu.SetKey(0xF, active)
				}
			}
		}

		if !paused {
			emu.Cycles(maxCycles)

			if emu.ShouldSound() && !muted {
				playSound()
			}
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

		tick2 := sdl.GetTicks()

		elapsed := tick2 - tick1
		remaining := interval - elapsed

		if elapsed < interval {
			sdl.Delay(remaining)
		}
	}
}
