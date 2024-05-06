package services

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/fs"

	"github.com/deezer/groroti/internal/staticEmbed"
	"github.com/goki/freetype/truetype"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func addLabel(img *image.RGBA, x, y int, label string, myFont *truetype.Font, size float64, col color.Color) {
	point := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}

	face := truetype.NewFace(myFont, &truetype.Options{
		Size: size,
	})

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  point,
	}
	d.DrawString(label)
}

func exportAsPNG(roti existingROTI) *image.RGBA {
	var y []int
	if roti.Description != "" {
		y = []int{200, 50, 90, 140, 175}
	} else {
		y = []int{150, 50, 0, 90, 125}
	}

	img := image.NewRGBA(image.Rect(0, 0, 1000, y[0]))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Create a sub-file system to load font
	staticFS, err := fs.Sub(staticEmbed.EmbeddedStatic, "static")
	if err != nil {
		log.Error().Err(err)
	}

	fontBytes, err := fs.ReadFile(staticFS, "Luciole-Regular.ttf")
	if err != nil {
		log.Error().Err(err)
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Error().Err(err)
	}

	addLabel(img, 5, y[1], fmt.Sprintf("ROTI - %d", roti.Id), font, 40, color.RGBA{200, 100, 0, 255})
	if roti.Description != "" {
		addLabel(img, 5, y[2], fmt.Sprintf("Meeting: %s", roti.Description), font, 32, color.Black)
	}
	addLabel(img, 5, y[3], fmt.Sprintf("Average ROTI: %0.2f | Min: %0.2f | Max: %0.2f", roti.Avg, roti.Min, roti.Max), font, 24, color.Black)
	addLabel(img, 5, y[4], fmt.Sprintf("Number of votes: %d", roti.NumVotes), font, 24, color.Black)

	return img
}

func exportAsCSV(roti existingROTI) (csv_strings []string) {
	csv_strings = []string{"ROTI ID,Description,Average ROTI,Min ROTI,Max ROTI,Number of Votes",
		fmt.Sprintf("%d,%s,%.2f,%.2f,%.2f,%d", roti.Id, roti.Description, roti.Avg, roti.Min, roti.Max, roti.NumVotes)}

	return csv_strings
}
