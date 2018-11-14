package resize

import (
	"fmt"
	"math"
)

// getCubic ...
func getCubic(oldImg MonoUInt8, toWidth, toHeight, ksize int) (xofs, yofs, alpha, beta []int, xmin, xmax int) {
	scaleX := float32(oldImg.Width) / float32(toWidth)
	scaleY := float32(oldImg.Height) / float32(toHeight)

	xmin, xmax = 0, toWidth
	for dx := 0; dx < toWidth; dx++ {
		fx := (float32(dx)+0.5)*scaleX - 0.5
		sx := int(fx)
		fx -= float32(sx)

		if sx < (ksize/2 - 1) {
			xmin = dx + 1
		}
		if sx+ksize/2 >= oldImg.Width {
			if xmax > dx {
				xmax = dx
			}
		}

		xofs = append(xofs, sx)
		cbuf := interpolateCubic(fx)

		for k := 0; k < ksize; k++ {
			alpha = append(alpha, int(math.Round(float64(cbuf[k]*2048.0))))
		}
	}

	for dy := 0; dy < toHeight; dy++ {
		fy := (float32(dy)+0.5)*scaleY - 0.5
		sy := int(fy)
		fy -= float32(sy)

		yofs = append(yofs, sy)
		cbuf := interpolateCubic(fy)

		for k := 0; k < ksize; k++ {
			beta = append(beta, int(math.Round(float64(cbuf[k]*2048.0))))
		}
	}
	return
}

// interpolateCubic ...
func interpolateCubic(x float32) (coeffs []float32) {
	const A float32 = -0.75
	coeffs = append(coeffs, ((A*(x+1.0)-5*A)*(x+1)+8*A)*(x+1)-4*A)
	coeffs = append(coeffs, ((A+2)*x-(A+3))*x*x+1)
	coeffs = append(coeffs, ((A+2)*(1-x)-(A+3))*(1-x)*(1-x)+1)
	coeffs = append(coeffs, 1.0-coeffs[0]-coeffs[1]-coeffs[2])
	return
}

// clip ...
func clip(x, a, b int) (out int) {
	var num int
	if x < b {
		num = x
	} else {
		num = b - 1
	}

	if x >= a {
		out = num
	} else {
		out = a
	}
	return
}

// getSrows ...
func getSrows(yof, imgHeight, ksize int) (srows []int) {
	ksize2 := ksize / 2
	for k := 0; k < ksize; k++ {
		sy := clip(yof-ksize2+1+k, 0, imgHeight)
		srows = append(srows, sy)
	}
	return
}

// hResizeCubic ...
func hResizeCubic(srcImg MonoUInt8, srows, xofs, alpha []int, dstWidth, xmin, xmax, ksize int) (result [][]int) {
	for k := 0; k < ksize; k++ {
		dx := 0
		limit := xmin
		alphaIdx := 0
		var row []int
		for {
			for ; dx < limit; dx++ {
				sx := xofs[dx] - 1
				v := 0
				for j := 0; j < 4; j++ {
					sxj := sx + j
					if sxj < 0 {
						sxj = 0
					}
					if sxj >= srcImg.Width {
						sxj = srcImg.Width - 1
					}
					pix := int(srcImg.Frame[srows[k]*srcImg.Width+sxj])
					v += pix * alpha[alphaIdx+j]
				}
				alphaIdx += 4
				row = append(row, v)
			}

			if limit == dstWidth {
				break
			}

			for ; dx < xmax; dx++ {
				sx := xofs[dx]
				pixIndex := srows[k]*srcImg.Width + sx - 1
				f1 := int(srcImg.Frame[pixIndex]) * alpha[alphaIdx]
				f2 := int(srcImg.Frame[pixIndex+1]) * alpha[alphaIdx+1]
				f3 := int(srcImg.Frame[pixIndex+2]) * alpha[alphaIdx+2]
				f4 := int(srcImg.Frame[pixIndex+3]) * alpha[alphaIdx+3]
				pix := f1 + f2 + f3 + f4
				row = append(row, pix)
				alphaIdx += 4
			}
			limit = dstWidth
		}
		result = append(result, row)
	}
	return
}

// vResizeCubic ...
func vResizeCubic(rows [][]int, beta []int, dstWidth int) (result []int, err error) {
	b0 := beta[0]
	b1 := beta[1]
	b2 := beta[2]
	b3 := beta[3]

	x, result, err := vecOpCubic(rows, beta, dstWidth)
	for ; x < dstWidth; x++ {
		val := castOpCubic(rows[0][x]*b0 + rows[1][x]*b1 + rows[2][x]*b2 + rows[3][x]*b3)
		result = append(result, val)
	}
	return
}

// vecOpCubic ...
func vecOpCubic(rows [][]int, beta []int, dstWidth int) (x int, result []int, err error) {
	x = 0
	scale := 1.0 / (2048.0 * 2048.0)
	b := arrayMul(beta, scale)

	// var s0, s1, f0, f1 [4]float64
	for ; x <= dstWidth-8; x += 8 {
		x0, x1, y0, y1 := rows[0][x:x+4], rows[0][x+4:x+8], rows[1][x:x+4], rows[1][x+4:x+8]
		s0 := arrayMul(x0, b[0])
		s1 := arrayMul(x1, b[0])
		f0 := arrayMul(y0, b[1])
		f1 := arrayMul(y1, b[1])

		s0, err = arrayAdd(s0, f0)
		if err != nil {
			return
		}
		s1, err = arrayAdd(s1, f1)
		if err != nil {
			return
		}

		x0, x1, y0, y1 = rows[2][x:x+4], rows[2][x+4:x+8], rows[3][x:x+4], rows[3][x+4:x+8]
		f0 = arrayMul(x0, b[2])
		f1 = arrayMul(x1, b[2])
		s0, err = arrayAdd(s0, f0)
		if err != nil {
			return
		}
		s1, err = arrayAdd(s1, f1)
		if err != nil {
			return
		}

		f0 = arrayMul(y0, b[3])
		f1 = arrayMul(y1, b[3])
		s0, err = arrayAdd(s0, f0)
		s1, err = arrayAdd(s1, f1)

		for _, val := range s0 {
			result = append(result, int(math.Round(val)))
		}
		for _, val := range s1 {
			result = append(result, int(math.Round(val)))
		}
	}

	return
}

// castOpCubic ...
func castOpCubic(val int) (out int) {
	const interResizeCofeBits int = 11
	SHIFT := interResizeCofeBits * 2
	DELTA := pow(2, (SHIFT - 1))
	out = (val + DELTA) / pow(2, SHIFT)
	return
}

// pow ...
func pow(x, n int) int {
	ret := 1
	for n != 0 {
		if n%2 != 0 {
			ret = ret * x
		}
		n /= 2
		x = x * x
	}
	return ret
}

// arrayAdd ...
func arrayAdd(arr1, arr2 []float64) (result []float64, err error) {
	if len(arr1) != len(arr2) {
		err = fmt.Errorf("In arrayAdd, the length of arr1 and arr2 must be equal")
		return
	}
	for i, val := range arr1 {
		result = append(result, val+arr2[i])
	}
	return
}

// arrayMul ...
func arrayMul(arr []int, factor float64) (result []float64) {
	for _, val := range arr {
		result = append(result, float64(val)*factor)
	}
	return
}
