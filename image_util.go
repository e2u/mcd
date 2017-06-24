package main

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"os"

	"github.com/e2u/goboot"
)

var (
	cssKeyWork = []string{
		`!`, `"`, `#`, `$`, `%`, `&`,
		`\`, `(`, `)`, `*`, `+`, `,`,
		`.`, `/`, `:`, `;`, `<`, `~`,
		`=`, `>`, `?`, `@`, `[`, `\`,
		`]`, `^`, "`", `{`, `|`, `}`, `-`}
)

type SpriteImage struct {
	Width    int
	Height   int
	Format   string
	Image    image.Image
	FileName string
}

// 读取多个图片文件并返回 SpriteImage 数组
func NewSpriteImageFromFiles(files []string) []*SpriteImage {
	var sis []*SpriteImage
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			goboot.Log.Errorf("open file error: %v", err.Error())
			continue
		}
		defer f.Close()
		i, format, err := image.Decode(f)
		if err != nil {
			goboot.Log.Errorf("decode image %v error: %v", file, err.Error())
			continue
		}
		sis = append(sis, &SpriteImage{
			Format:   format,
			Image:    i,
			Width:    i.Bounds().Max.X,
			Height:   i.Bounds().Max.Y,
			FileName: file,
		})
	}
	return sis
}
