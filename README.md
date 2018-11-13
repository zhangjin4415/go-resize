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

```
imgResized := resize.Resize(srcImg, toWidth, toHeight, resize.InterCubic)
```

## It's developing, now it's only can resize by InterCubic!