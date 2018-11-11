package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// // FrameUInt8 ...
// type FrameUInt8 []uint8

// // MonoUInt8 ...
// type MonoUInt8 struct {
// 	Frame  FrameUInt8
// 	Width  int
// 	Height int
// }

func main() {
	imgdata := getimg()

	//部分值由于float影响 略有差异
	xofs, ialpha, yofs, ibeta, xmin, xmax, ksize := getcanshu(imgdata, 50, 50)

	//resizeGeneric_ ..

	out:=invoker(imgdata,xofs,yofs,ialpha,ibeta,50,50,xmin, xmax, ksize)


	fmt.Println(out)
	// fmt.Println(xofs, ialpha, yofs, ibeta, xmin, xmax, ksize)
}

// invoker ..
func invoker(img MonoUInt8, xofs,yofs, ialpha,ibeta []int,width,height,xmin, xmax,ksize int) (result MonoUInt8) {

	result.Width=width
	result.Height=height

	const MAXESIZE int = 16
	// const WIDTH int = 50
	srcWidth, dstWidth:=img.Width,50///

	var prev_sy[MAXESIZE] int
	
	for k := 0; k < ksize; k++ {
		prev_sy[k] = -1;
		// rows[k] = (WT*)_buffer + bufstep*k;
	}

	ibetaIdx:=0
	for dy := 0; dy < height; dy++{//, beta += ksize
		sy0 := yofs[dy]
		// k0:=ksize
		// k1:=0
		ksize2 := ksize/2;
		// fmt.Println(sy0," ",k0," ",k1," ",ksize2)

		var srows[4][] int
		for k := 0; k < ksize; k++ {
			sy:=clip(sy0 - ksize2 + 1 + k, 0, img.Height)
			// fmt.Println(sy)

			for _,val := range img.Frame[sy*img.Width:(sy+1)*img.Width]{
				srows[k]=append(srows[k],int(val))
			}
		}

		// fmt.Println(srows)

		rows:=hresize(srows,xofs,ialpha,srcWidth, dstWidth,1,xmin, xmax,ksize)//cn=1
		out:=vresize(rows,ibeta[ibetaIdx:ibetaIdx+ksize],dstWidth)

		ibetaIdx+=4
		
		for _,val := range out{
			result.Frame=append(result.Frame,uint8(val))
		}
		
		// fmt.Println(out)
		// fmt.Println(srows,"----------")
	}

	return
}

//
func vresize(rows [4][50]int, beta []int, dstWidth int)(out []int){
	b0 := beta[0]
	b1 := beta[1]
	b2 := beta[2]
	b3 := beta[3]

	x,result := vecOp(rows, beta, dstWidth)
	for _,r := range result{
		out=append(out,r)
	}

	for ;x<dstWidth;x++{
		val:=castOp(rows[0][x]*b0+rows[1][x]*b1+rows[2][x]*b2+rows[3][x]*b3)
		out=append(out,val)
	}
	return
}

func vecOp(rows [4][50]int, beta []int, dstWidth int)(x int,result []int){
	x = 0
	scale := 1.0/(2048.0*2048.0)
	b0:=float64(beta[0])*scale
	b1:=float64(beta[1])*scale
	b2:=float64(beta[2])*scale
	b3:=float64(beta[3])*scale
	
	var s0, s1, f0, f1 [4]float64
	for ;x<=dstWidth-8;x+=8{
		x0:=rows[0][x:x+4]
		x1:=rows[0][x+4:x+8]
		y0:=rows[1][x:x+4]
		y1:=rows[1][x+4:x+8]

		s0[0]=float64(x0[0])*b0
		s0[1]=float64(x0[1])*b0
		s0[2]=float64(x0[2])*b0
		s0[3]=float64(x0[3])*b0

		s1[0]=float64(x1[0])*b0
		s1[1]=float64(x1[1])*b0
		s1[2]=float64(x1[2])*b0
		s1[3]=float64(x1[3])*b0

		f0[0]=float64(y0[0])*b1
		f0[1]=float64(y0[1])*b1
		f0[2]=float64(y0[2])*b1
		f0[3]=float64(y0[3])*b1

		f1[0]=float64(y1[0])*b1
		f1[1]=float64(y1[1])*b1
		f1[2]=float64(y1[2])*b1
		f1[3]=float64(y1[3])*b1
		
		s0[0]+=f0[0]
		s0[1]+=f0[1]
		s0[2]+=f0[2]
		s0[3]+=f0[3]

		s1[0]+=f1[0]
		s1[1]+=f1[1]
		s1[2]+=f1[2]
		s1[3]+=f1[3]

		//
		x0 =rows[2][x:x+4]
		x1 =rows[2][x+4:x+8]
		y0 =rows[3][x:x+4]
		y1 =rows[3][x+4:x+8]

		f0[0]=float64(x0[0])*b2
		f0[1]=float64(x0[1])*b2
		f0[2]=float64(x0[2])*b2
		f0[3]=float64(x0[3])*b2

		f1[0]=float64(x1[0])*b2
		f1[1]=float64(x1[1])*b2
		f1[2]=float64(x1[2])*b2
		f1[3]=float64(x1[3])*b2

		s0[0]+=f0[0]
		s0[1]+=f0[1]
		s0[2]+=f0[2]
		s0[3]+=f0[3]

		s1[0]+=f1[0]
		s1[1]+=f1[1]
		s1[2]+=f1[2]
		s1[3]+=f1[3]

		f0[0]=float64(y0[0])*b3
		f0[1]=float64(y0[1])*b3
		f0[2]=float64(y0[2])*b3
		f0[3]=float64(y0[3])*b3

		f1[0]=float64(y1[0])*b3
		f1[1]=float64(y1[1])*b3
		f1[2]=float64(y1[2])*b3
		f1[3]=float64(y1[3])*b3

		s0[0]+=f0[0]
		s0[1]+=f0[1]
		s0[2]+=f0[2]
		s0[3]+=f0[3]

		s1[0]+=f1[0]
		s1[1]+=f1[1]
		s1[2]+=f1[2]
		s1[3]+=f1[3]

		
		for _,val := range s0{
			result=append(result,int(math.Round(val)))
		}
		for _,val := range s1{
			result=append(result,int(math.Round(val)))
		}
	}
	
	return
}

func castOp(val int)(out int){
	const INTER_RESIZE_COEF_BITS int = 11
	SHIFT := INTER_RESIZE_COEF_BITS*2
	
	// DELTA := 1 << (INTER_RESIZE_COEF_BITS-1)
	DELTA := pow(2,(SHIFT-1))
	// out=(val + DELTA)>>SHIFT
	out=(val + DELTA)/pow(2,SHIFT)////////////////////////////// 
	return
}

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




// hresize ...
func hresize(srows [4][]int,xofs, alpha []int,srcWidth, dstWidth, cn, xmin, xmax, ksize int)(rows [4][50]int){
	for k := 0; k < ksize; k++ {
		dx := 0
		limit := xmin;
		alphaIdx:=0
		for{
			for ;dx<limit;dx++{/////////// dst[0] dst[49]
				sx:=xofs[dx]-cn;
				v:=0
				for j:=0;j<4;j++{
					sxj:=sx+j*cn
					if sxj<0 || sxj>=srcWidth{
						for{
							if sxj<0{
								sxj+=cn;
							}else{
								break
							}
						}
						for{
							if sxj>=srcWidth{
								sxj-=cn
							}else{
								break
							}
						}
					}
					// fmt.Println(sxj)
					v+=srows[k][sxj]*alpha[alphaIdx+j]
				}
				alphaIdx+=4
				rows[k][dx]=v/////////////////////////////////
			}

			/////////////////////
			if limit==dstWidth{
				break
			}

			//////
			for ;dx<xmax;dx++{
				sx:=xofs[dx]
				rows[k][dx]=srows[k][sx-cn]*alpha[alphaIdx]+srows[k][sx]*alpha[alphaIdx+1]+srows[k][sx+cn]*alpha[alphaIdx+2]+srows[k][sx+cn*2]*alpha[alphaIdx+3]
				alphaIdx+=4
			}
			limit = dstWidth;
		}
	}
	return
}


func clip(x, a, b int)(out int){
	var num int
	if x<b{
		num=x
	}else{
		num=b-1
	}

	if x>=a{
		out=num
	}else{
		out=a
	}
	// return x >= a ? (x < b ? x : b-1) : a;
	return
}

// getcanshu ...
func getcanshu(img MonoUInt8, width, height int) (xofs, ialpha, yofs, ibeta []int, xmin, xmax, ksize int) {
	var k, sx, sy, dx, dy int
	var scaleX, scaleY, fx, fy float32

	//
	ksize = 4
	ksize2 := ksize / 2
	// scale_x = width / img.Width
	// scale_y = height / img.Height

	cn := 1 ////////////////////////////////

	xmin = 0
	xmax = width
	width = width * cn

	const MAXESIZE int = 16
	var cbuf [MAXESIZE]float32

	scaleX = float32(img.Width) / float32(width)
	scaleY = float32(img.Height) / float32(height)

	// fmt.Println(scaleX, scaleY)

	//for width ...
	for dx = 0; dx < width; dx++ {
		// 原来是 if ... else ...
		fx = (float32(dx)+0.5)*scaleX - 0.5 ///////提取整数sx 和 小数fx 部分
		sx = int(fx)
		fx -= float32(sx)

		if sx < (ksize2 - 1) {
			xmin = dx + 1
		}

		if sx+ksize2 >= img.Width {
			if xmax > dx { //max取 最小 dx
				xmax = dx
			}
		}

		for k = 0; k < cn; k++ {
			sx *= cn
			// xofs[dx*cn+k] = sx + k
			xofs = append(xofs, sx+k)
		}

		//INTER_CUBIC
		cbuf = interpolateCubic(fx)
		// if dx < 7 {
		// 	for i := 0; i < 4; i++ {
		// 		fmt.Println("---", cbuf[i])
		// 	}
		// }

		// if( fixpt ) ...
		for k = 0; k < ksize; k++ {
			//四舍五入
			// ialpha[dx*cn*ksize+k] = int(math.round(float64(cbuf[k]) * 2048.0))
			ialpha = append(ialpha, int(math.Round(float64(cbuf[k]*2048.0))))
		}
		for ; k < cn*ksize; k++ { //not run
			// ialpha[dx*cn*ksize+k] = ialpha[dx*cn*ksize+k-ksize]
			ialpha = append(ialpha, ialpha[dx*cn*ksize+k-ksize])
		}
	}

	//for height ...
	for dy = 0; dy < height; dy++ {
		// 原来是 if ... else ...
		fy = (float32(dy)+0.5)*scaleY - 0.5
		sy = int(fy)
		fy -= float32(sy)

		yofs = append(yofs, sy)

		//INTER_CUBIC ...
		cbuf = interpolateCubic(fy)

		//fixpt
		for k = 0; k < ksize; k++ {
			ibeta = append(ibeta, int(math.Round(float64(cbuf[k]*2048.0))))
		}
	}

	// fmt.Println(k, sx, sy, dx, dy, fx, fy, cbuf)
	return
}

func interpolateCubic(x float32) (coeffs [16]float32) {
	const A float32 = -0.75
	// coeffs = make([16]float32)
	coeffs[0] = ((A*(x+1.0)-5*A)*(x+1)+8*A)*(x+1) - 4*A
	coeffs[1] = ((A+2)*x-(A+3))*x*x + 1
	coeffs[2] = ((A+2)*(1-x)-(A+3))*(1-x)*(1-x) + 1
	coeffs[3] = 1.0 - coeffs[0] - coeffs[1] - coeffs[2]
	return
}

// getimg ...
func getimg() (out MonoUInt8) {
	f, err := os.Open("im.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n') //以'\n'为结束符读入一行

		if err != nil || io.EOF == err {
			break
		}
		// 去除空格
		line = strings.Replace(line, " ", "", -1)
		line = strings.Split(line, ";")[0]
		data := strings.Split(line, ",")
		// fmt.Println(data)

		out.Width = len(data)
		out.Height++
		for _, d := range data {
			u, _ := strconv.Atoi(d)
			out.Frame = append(out.Frame, uint8(u))
		}
	}
	return
}
