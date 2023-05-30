package robot

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"os"
)

func square(path string) string {
	// 打开图片文件
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 解码图像
	img, _, err := image.Decode(file)

	// 计算需要填充的宽度和高度
	size := img.Bounds().Size()
	max := size.X
	if size.Y > max {
		max = size.Y
	}
	dx := (max - size.X) / 2
	dy := (max - size.Y) / 2

	// 创建新的PNG图片文件
	path = fmt.Sprintf("/tmp/square%s.jpg", GenerateToken(32))
	out, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// 创建新的image.RGBA类型的对象
	rgba := image.NewRGBA(image.Rect(0, 0, max, max))

	// 将原始图片绘制到新的image.RGBA对象中心位置
	draw.Draw(rgba, rgba.Bounds(), image.Transparent, image.Point{}, draw.Src)
	draw.Draw(rgba, image.Rect(dx, dy, dx+size.X, dy+size.Y), img, image.Point{0, 0}, draw.Src)

	// 保存为PNG格式文件
	err = png.Encode(out, rgba)
	if err != nil {
		panic(err)
	}

	return path
}
