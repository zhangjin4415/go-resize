# go-resize is the opencv's resize for GO

Image is defined by MonoUInt8

```
// FrameUInt8 ...
type FrameUInt8 []uint8

// MonoUInt8 ...
type MonoUInt8 struct {
	Frame  FrameUInt8
	Width  int
	Height int
}
```
## use
### go get github.com/zhangjin4415/go-resize

```
imgResized := resize.Resize(srcImg, toWidth, toHeight, resize.InterCubic)
```

## now it's only can resize gray image by InterCubic!

## It's developing... 