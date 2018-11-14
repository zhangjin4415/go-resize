/*
Copyright (c) 2018, Zhang Jin <1654637062@qq.com>
*/

// Example:
//     imgResized := resize.Resize(srcImg, toWidth, toHeight, resize.InterCubic)

package resize

import (
	"fmt"
	"math"
)

// Interpolation ...
type Interpolation int

// interpolation algorithm
const (
	/** nearest neighbor interpolation */
	InterNearest Interpolation = iota
	/** bilinear interpolation */
	InterLinear
	/** bicubic interpolation */
	InterCubic
)

// Resize ...
func Resize(srcImg MonoUInt8, toWidth, toHeight int, interp Interpolation) (dstImg MonoUInt8, err error) {
	if toWidth <= 0 {
		err = fmt.Errorf("toWidth must be > 0")
		return
	}
	if toHeight <= 0 {
		err = fmt.Errorf("toHeight must be > 0")
		return
	}
	if len(srcImg.Frame) == 0 {
		err = fmt.Errorf("input oldImg has no pixels")
		return
	}
	if toWidth == srcImg.Width && toHeight == srcImg.Height {
		dstImg = srcImg
		return
	}

	// interp
	switch interp {
	case InterNearest:
		err = fmt.Errorf("InterNearest is developing")
		return

	case InterLinear:
		dstImg, err = resizeLinear(srcImg, toWidth, toHeight)
		return

	case InterCubic:
		dstImg, err = resizeCubic(srcImg, toWidth, toHeight)
		return
	}
	return
}


// resizeLinear ...
func resizeLinear(srcImg MonoUInt8, toWidth, toHeight int) (dstImg MonoUInt8, err error) {
	scaleX := float32(srcImg.Width) / float32(toWidth)
	scaleY := float32(srcImg.Height) / float32(toHeight)
	for j := 0; j < toHeight; j++ {
		fy := (float32(j)+0.5)*scaleY - 0.5
		sy := int(math.Floor(float64(fy)))
		fy -= float32(sy)
		if sy > srcImg.Height {
			sy = srcImg.Height - 1
		}
		if sy < 0 {
			sy = 0
		}
		var cbufy [2]float32
		cbufy[0] = (1.0 - fy) * 2048
		cbufy[1] = 2048 - cbufy[0]

		for i := 0; i < toWidth; i++ {
			fx := (float32(i)+0.5)*scaleX - 0.5
			sx := int(math.Floor(float64(fx)))
			fx -= float32(sx)

			if sx < 0 {
				fx, sx = 0, 0
			}
			if sx > srcImg.Width {
				fx, sx = 0, srcImg.Width-1
			}

			var cbufx [2]float32
			cbufx[0] = (1.0 - fx) * 2048
			cbufx[1] = 2048 - cbufx[0]

			if sx<srcImg.Width-1 && sy<srcImg.Height-1{
				data1 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[0] * cbufy[0]
				data2 := float32(srcImg.Frame[(sy+1)*srcImg.Width+sx]) * cbufx[0] * cbufy[1]
				data3 := float32(srcImg.Frame[sy*srcImg.Width+sx+1]) * cbufx[1] * cbufy[0]
				data4 := float32(srcImg.Frame[(sy+1)*srcImg.Width+sx+1]) * cbufx[1] * cbufy[1]
				data := (data1 + data2 + data3 + data4) / float32(pow(2, 22))
				dstImg.Frame = append(dstImg.Frame, uint8(math.Round(float64(data))))
			}else if sx<srcImg.Width-1{
				data1 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[0] * cbufy[0]
				data2 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[0] * cbufy[1]
				data3 := float32(srcImg.Frame[sy*srcImg.Width+sx+1]) * cbufx[1] * cbufy[0]
				data4 := float32(srcImg.Frame[sy*srcImg.Width+sx+1]) * cbufx[1] * cbufy[1]
				data := (data1 + data2 + data3 + data4) / float32(pow(2, 22))
				dstImg.Frame = append(dstImg.Frame, uint8(math.Round(float64(data))))
			}else if sy<srcImg.Height-1{
				data1 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[0] * cbufy[0]
				data2 := float32(srcImg.Frame[(sy+1)*srcImg.Width+sx]) * cbufx[0] * cbufy[1]
				data3 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[1] * cbufy[0]
				data4 := float32(srcImg.Frame[(sy+1)*srcImg.Width+sx]) * cbufx[1] * cbufy[1]
				data := (data1 + data2 + data3 + data4) / float32(pow(2, 22))
				dstImg.Frame = append(dstImg.Frame, uint8(math.Round(float64(data))))
			}else{
				data1 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[0] * cbufy[0]
				data2 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[0] * cbufy[1]
				data3 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[1] * cbufy[0]
				data4 := float32(srcImg.Frame[sy*srcImg.Width+sx]) * cbufx[1] * cbufy[1]
				data := (data1 + data2 + data3 + data4) / float32(pow(2, 22))
				dstImg.Frame = append(dstImg.Frame, uint8(math.Round(float64(data))))
			}
			
		}

	}
	dstImg.Width = toWidth
	dstImg.Height = toHeight
	return
}

// resizeCubic ...
func resizeCubic(srcImg MonoUInt8, toWidth, toHeight int) (dstImg MonoUInt8, err error) {
	const ksize int = 4
	xofs, yofs, alpha, beta, xmin, xmax := getCubic(srcImg, toWidth, toHeight, ksize)

	var betaIdx int
	for dy := 0; dy < toHeight; dy++ {
		srows := getSrows(yofs[dy], srcImg.Height, ksize) //0,0,1,2  2,3,4,5 ...
		rows := hResizeCubic(srcImg, srows, xofs, alpha, toWidth, xmin, xmax, ksize)
		out, e := vResizeCubic(rows, beta[betaIdx:betaIdx+ksize], toWidth)
		if e != nil {
			err = e
			break
		}
		betaIdx += 4
		for _, val := range out {
			dstImg.Frame = append(dstImg.Frame, uint8(val))
		}
		dstImg.Width, dstImg.Height = toWidth, toHeight
	}
	return
}
