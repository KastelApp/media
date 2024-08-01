
package utils

import (
	"github.com/disintegration/imaging"
)

func ResizeImage(filePath string, width, height int) (string, error) {
	src, err := imaging.Open(filePath)
	if err != nil {
		return "", err
	}

	dst := imaging.Resize(src, width, height, imaging.Lanczos)

	outputPath := filePath + "_resized.png"
	err = imaging.Save(dst, outputPath)
	if err != nil {
		return "", err
	}

	return outputPath, nil
}
