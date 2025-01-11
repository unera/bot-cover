package main

import (
	_ "embed"
	"fmt"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func annotateImage(mw *imagick.MagickWand,
	profile *Profile, cfg *Config,
	il *ImageLabel, gravity imagick.GravityType) {

	if len(il.Text) == 0 {
		return
	}
	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

	mw2 := imagick.NewMagickWand()
	defer mw2.Destroy()

	var fontSize float64
	if width > height {
		fontSize = float64(height) * float64(il.Size) / 100
	} else {
		fontSize = float64(width) * float64(il.Size) / 100
	}

	dw := imagick.NewDrawingWand()
	defer dw.Destroy()

	pw := imagick.NewPixelWand()
	defer pw.Destroy()

	pw.SetColor("none")
	mw2.NewImage(width, height, pw)
	if err := mw2.SetImageFormat("png"); err != nil {
		panic(err)
	}

	pw.SetColor(il.Color)
	dw.SetFillColor(pw)

	pw.SetColor(il.StrokeColor)
	dw.SetStrokeColor(pw)

	dw.SetStrokeWidth(fontSize / 80)
	dw.SetGravity(gravity)
	dw.SetFontSize(fontSize)
	if fontFile, ok := cfg.App.Fonts[il.Font]; ok {
		if err := dw.SetFont(fontFile); err != nil {
			panic(fmt.Sprintf("Can not set font: %s", err))
		}
	}
	dw.SetTextAntialias(true)

	dw.Annotation(0, fontSize/2, il.Text)

	mw2.DrawImage(dw)

	mw.SetLastIterator()
	mw.AddImage(mw2)
}

// MakeImage apply text to image
func MakeImage(raw []byte, profile *Profile, cfg *Config) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImageBlob(raw); err != nil {
		panic(err)
	}

	width, height := mw.GetImageWidth(), mw.GetImageHeight()

	mwo := imagick.NewMagickWand()
	defer mwo.Destroy()

	pw := imagick.NewPixelWand()
	defer pw.Destroy()
	pw.SetColor("none")
	mwo.NewImage(width, height, pw)
	if err := mwo.SetImageFormat("png"); err != nil {
		panic(err)
	}

	mwo.SetFirstIterator()
	mwo.AddImage(mw)
	mwo.SetLastIterator()
	mwo.RemoveImage()

	annotateImage(mwo, profile, cfg, &profile.Image.Top, imagick.GRAVITY_NORTH)
	annotateImage(mwo, profile, cfg, &profile.Image.Bottom, imagick.GRAVITY_SOUTH)

	mwo.ResetIterator()
	mwe := mwo.MergeImageLayers(imagick.IMAGE_LAYER_COMPOSITE)
	defer mwe.Destroy()

	if blob, err := mwe.GetImagesBlob(); err != nil {
		panic(err)
	} else {
		return blob
	}
}

// MakePredefinedImage apply text to predefined image
func MakePredefinedImage(profile *Profile, cfg *Config) []byte {

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImageBlob(predefinedImage); err != nil {
		panic(err)
	}

	if err := mw.SetImageFormat("png"); err != nil {
		panic(err)
	}

	if err := mw.CropImage(
		uint(profile.Image.Width),
		uint(profile.Image.Height),
		0,
		0); err != nil {
		panic(err)
	}
	if err := mw.SetImagePage(
		uint(profile.Image.Width),
		uint(profile.Image.Height),
		0,
		0); err != nil {
		panic(err)
	}

	mw.SetSize(uint(profile.Image.Width), uint(profile.Image.Height))

	if blob, err := mw.GetImageBlob(); err != nil {
		panic(err)
	} else {
		return MakeImage(blob, profile, cfg)
	}
}

//go:embed example.png
var predefinedImage []byte

func iniitImageSystem() {
	imagick.Initialize()
}

func closeImageSystem() {
	imagick.Terminate()
}
