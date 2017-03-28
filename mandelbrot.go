package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math/cmplx"
	"os"
)

var (
	pointC = make(chan point)
	stopC  = make(chan struct{})
)

type point struct {
	X int
	Y int
	N int
}

const maxCount int = 2048

func mandelbrot(x, y float64) int {
	c := complex(x, y)
	z := c
	for i := 0; i < maxCount; i++ {
		if cmplx.Abs(z) > 2 {
			return i + 1
		}
		z = z*z + c
	}
	return maxCount
}

func gen(min, max float64, num int) <-chan float64 {
	out := make(chan float64)
	go func() {
		step := (max - min) / float64(num)
		for i := min; i < max; i += step {
			out <- i
		}
		close(out)
	}()
	return out
}

func drawImg(width, height int) {
	out, err := os.Create("mandelbrot1.jpg")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	imgRect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(imgRect)
	for p := range pointC {
		img.Set(p.X, p.Y, calcColor(p.N))
	}
	err = jpeg.Encode(out, img, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	stopC <- struct{}{}
}

func calcColor(n int) color.Color {
	return color.RGBA{uint8((n + 50) % 256), uint8(n % 256), uint8(n % 256), 255}
}

func main() {
	var width, height = 1000, 1000
	go drawImg(width, height)
	var x, y int
	for i := range gen(-2.0, 0.5, width) {
		y = 0
		for j := range gen(-1.25, 1.25, height) {
			pointC <- point{x, y, mandelbrot(i, j)}
			y++
		}
		x++
	}
	close(pointC)
	<-stopC
}
