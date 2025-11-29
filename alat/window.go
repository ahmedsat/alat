package alat

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"time"

	"github.com/ahmedsat/noor/gl"
	_ "github.com/ahmedsat/noor/gl"
	"github.com/ahmedsat/noor/window"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/skip2/go-qrcode"
)

type Window struct {
	window.Window
	QrTex gl.TextureId
	Show  func()
}

type NonThreadSafe struct {
	Func func() error
	Err  chan error
}

type WindowCreator struct {
	Windows       map[string]*Window
	NonThreadSafe chan NonThreadSafe
}

func NewWindowCreator() *WindowCreator {
	return &WindowCreator{
		map[string]*Window{},
		make(chan NonThreadSafe),
	}
}

func (wc *WindowCreator) NonThreadSafeExec(nts NonThreadSafe) error {
	wc.NonThreadSafe <- nts
	return <-nts.Err
}

func (wc *WindowCreator) Show() {
	for {
		select {
		case nts := <-wc.NonThreadSafe:
			err := nts.Func()
			nts.Err <- err
		default:

			for id, w := range wc.Windows {
				if w.ShouldClose() {
					w.Destroy()
					delete(wc.Windows, id)

					continue
				}
				if glfw.Press == w.GetKey(glfw.KeyEscape) {
					wc.Close(id, nil)
					break
				}

				if w.Show != nil {
					w.Show()
				}

				w.SwapBuffers()
				glfw.PollEvents()
			}

			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (wc *WindowCreator) Close(id string, reply *string) error {
	w, ok := wc.Windows[id]
	if !ok {
		*reply = fmt.Sprintf("Window with id %s does not exist", id)
		return nil
	}

	w.SetShouldClose(true)
	return nil
}

type SolidColorArgs struct {
	Id     string
	Width  int
	Height int
	Title  string
	Color  color.RGBA
}

func (Args *SolidColorArgs) validate() error {

	var err error

	if Args.Width <= 0 {
		err = errors.Join(err, fmt.Errorf("Width must be positive"))
	}
	if Args.Height <= 0 {
		err = errors.Join(err, fmt.Errorf("Height must be positive"))
	}

	if Args.Title == "" {
		err = errors.Join(err, fmt.Errorf("Title must be not empty"))
	}

	if Args.Id == "" {
		err = errors.Join(err, fmt.Errorf("Id must be not empty"))
	}

	return err
}

func (wc *WindowCreator) Solid(Args SolidColorArgs, result *string) error {
	*result = ""

	err := Args.validate()
	if err != nil {
		*result = err.Error()
		return nil
	}

	show := func() {
		gl.ClearColor(Args.Color)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}

	if wc.Windows[Args.Id] != nil {
		*result = fmt.Sprintf("Window with id %s already exists and will be overwritten", Args.Id)
		wc.Windows[Args.Id].SetTitle(Args.Title)
		wc.Windows[Args.Id].SetSize(Args.Width, Args.Height)
		wc.Windows[Args.Id].Show = show
		return nil
	}

	wc.Windows[Args.Id] = &Window{}
	err = wc.NonThreadSafeExec(NonThreadSafe{
		Func: func() error {
			wc.Windows[Args.Id].Window, err = window.NewWindow(Args.Width, Args.Height, Args.Title)
			if err != nil {
				return err
			}
			wc.Windows[Args.Id].Show = show
			return nil
		},
		Err: make(chan error),
	})
	if err != nil {
		return err
	}

	return nil
}

type QrArgs struct {
	Id            string
	Title         string
	Text          string
	RecoveryLevel int
	Size          int
}

func (Args *QrArgs) validate() error {

	var err error

	if Args.Title == "" {
		err = errors.Join(err, fmt.Errorf("Title must be not empty"))
	}

	if Args.Text == "" {
		err = errors.Join(err, fmt.Errorf("Text must be not empty"))
	}

	if Args.Id == "" {
		err = errors.Join(err, fmt.Errorf("Id must be not empty"))
	}

	if Args.RecoveryLevel < 0 || Args.RecoveryLevel > 3 {
		err = errors.Join(err, fmt.Errorf("RecoveryLevel must be in range [0-3]"))
	}

	if Args.Size <= 0 {
		err = errors.Join(err, fmt.Errorf("Size must be positive"))
	}

	return err
}

func (wc *WindowCreator) Qr(Args QrArgs, result *string) error {

	var sh gl.Shader
	var vao gl.VertexArrayId

	*result = ""

	err := Args.validate()
	if err != nil {
		*result = err.Error()
		return nil
	}

	if window := wc.Windows[Args.Id]; window != nil {
		*result = fmt.Sprintf("Window with id %s already exists", Args.Id)
		wc.Windows[Args.Id].SetTitle(Args.Title)
		wc.Windows[Args.Id].SetSize(Args.Size, Args.Size)

		err := wc.NonThreadSafeExec(NonThreadSafe{
			Func: func() error {
				newTex, err := qrToTexture(Args.Text, Args.Size, qrcode.RecoveryLevel(Args.RecoveryLevel))
				if err != nil {
					return err
				}
				newTex, window.QrTex = window.QrTex, newTex
				newTex.Delete()

				return nil
			},
			Err: make(chan error),
		})

		return err
	}

	var vertices = []float32{
		1, 1, 1, 0, // top right
		1, -1, 1, 1, // bottom right
		-1, -1, 0, 1, // bottom left
		-1, 1, 0, 0, // top left
	}
	var indices = []uint32{ // note that we start from 0!
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}

	wc.Windows[Args.Id] = &Window{}

	err = wc.NonThreadSafeExec(NonThreadSafe{
		Func: func() error {
			wc.Windows[Args.Id].Window, err = window.NewWindow(Args.Size, Args.Size, Args.Title)
			if err != nil {
				return err
			}

			sh, err = gl.NewShader("shaders/qr.vert", "shaders/qr.frag")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			vao = gl.GenVertexArray()
			vao.Bind()

			vbo := gl.GenBuffer()
			vbo.ArrayBufferData(vertices, gl.STATIC_DRAW)

			ebo := gl.GenBuffer()
			ebo.ElementBufferData(indices, gl.STATIC_DRAW)

			gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 4*4, 0)
			gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 4*4, 2*4)

			wc.Windows[Args.Id].QrTex, err = qrToTexture(Args.Text, Args.Size, qrcode.RecoveryLevel(Args.RecoveryLevel))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
		Err: make(chan error),
	})
	if err != nil {
		return err
	}

	wc.Windows[Args.Id].Show = func() {
		wc.Windows[Args.Id].QrTex.Bind()

		sh.Use()
		vao.Bind()

		gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, 0)
	}

	return nil
}

func qrToTexture(text string, size int, level qrcode.RecoveryLevel) (tex gl.TextureId, err error) {
	var pngBytes []byte
	pngBytes, err = qrcode.Encode(text, level, size)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	img, err := png.Decode(bytes.NewBuffer(pngBytes))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tex, err = gl.NewTextureFromImage(img)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}
