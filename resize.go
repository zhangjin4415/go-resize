/*
Copyright (c) 2018, Zhang Jin <1654637062@qq.com>
*/

// Example:
//     imgResized := resize.Resize(srcImg, toWidth, toHeight, resize.InterCubic)

package main

import (
	"fmt"
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
		err = fmt.Errorf("InterLinear is developing")
		return

	case InterCubic:
		dstImg, err = resizeCubic(srcImg, toWidth, toHeight)
		return
	}
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

	fmt.Println(xofs, yofs, alpha, beta, xmin, xmax)
	return
}
