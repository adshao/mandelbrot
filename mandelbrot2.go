package main

import (
    "fmt"
    "image"
    "image/color"
    "image/jpeg"
    "math/cmplx"
    "os"
    "sync"
)

type point struct {
    X int
    Y int
    N int
}

const maxCount int = 2048

func mandelbrot(r, i float64) int {
    c := complex(r, i)
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

func drawImg(img *image.RGBA, pointC chan point) {
    for p := range pointC {
        img.Set(p.X, p.Y, calcColor(p.N))
    }
}

func calcColor(n int) color.Color {
    return color.RGBA{uint8(n % 256), uint8((n + 50) % 256), uint8(n % 256), 255}
}

func main() {
    out, err := os.Create("mandelbrot2.jpg")
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    var width, height = 1000, 1000
    xMin, xMax := -2.0, 0.5
    yMin, yMax := -1.25, 1.25
    xNum, yNum := 3, 2

    widthStep := width / xNum
    heightStep := height / yNum
    xStep := (xMax - xMin) / float64(xNum)
    yStep := (yMax - yMin) / float64(yNum)

    imgRect := image.Rect(0, 0, width, height)
    img := image.NewRGBA(imgRect)

    var wg sync.WaitGroup
    for xi := 0; xi < xNum; xi++ {
        for yi := 0; yi < yNum; yi++ {
            wg.Add(1)
            subImg := img.SubImage(
                image.Rect(xi * widthStep, yi * heightStep, 
                    (xi + 1) * widthStep, (yi + 1) * heightStep)).(*image.RGBA)
            pointC := make(chan point)
            go drawImg(subImg, pointC)
            go func(xi, yi int, pointC chan point) {
                defer wg.Done()
                x := xi * widthStep
                var xx, yx float64
                if xi == xNum -1 {
                    xx = xMax
                } else {
                    xx = xMin + float64(xi + 1) * xStep
                }
                if yi == yNum - 1 {
                    yx = yMax
                } else {
                    yx = yMin + float64(yi + 1) * yStep
                }
                for i := range gen(xMin + float64(xi) * xStep, xx, widthStep) {
                    y := yi * heightStep
                    for j := range gen(yMin + float64(yi) * yStep, yx, heightStep) {
                        pointC <- point{x, y, mandelbrot(i, j)}
                        y++
                    }
                    x++
                }
                close(pointC)
            }(xi, yi, pointC)
        }
    }
    wg.Wait()

    err = jpeg.Encode(out, img, nil)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
