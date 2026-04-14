package service

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"

	// Регистрируем декодер PNG, чтобы image.Decode умел читать PNG-файлы.
	_ "image/png"
	"imageProcessor/internal/dto"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
)

// ProcessMessage выполняет фоновую обработку изображения.
func (s *Service) ProcessMessage(ctx context.Context, msg dto.ProcessImageMessage) error {
	if err := s.repo.UpdateImageStatus(ctx, msg.ID, "processing", ""); err != nil {
		return err
	}

	srcFile, err := os.Open(msg.OriginalPath)
	if err != nil {
		_ = s.repo.UpdateImageStatus(ctx, msg.ID, "failed", err.Error())
		return err
	}
	defer srcFile.Close()

	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		_ = s.repo.UpdateImageStatus(ctx, msg.ID, "failed", err.Error())
		return err
	}

	resultImg := srcImg

	if msg.Resize {
		resultImg = resize.Resize(500, 0, resultImg, resize.Lanczos3)
	}

	if msg.Watermark {
		resultImg = addSimpleWatermark(resultImg)
	}

	processedPath := filepath.Join(s.storage.ProcessedDir(), msg.ID+".jpg")
	if err := saveJPEG(processedPath, resultImg); err != nil {
		_ = s.repo.UpdateImageStatus(ctx, msg.ID, "failed", err.Error())
		return err
	}

	var thumbPath string
	if msg.Thumb {
		thumbImg := resize.Thumbnail(200, 200, resultImg, resize.Lanczos3)
		thumbPath = filepath.Join(s.storage.ThumbsDir(), msg.ID+".jpg")

		if err := saveJPEG(thumbPath, thumbImg); err != nil {
			_ = s.repo.UpdateImageStatus(ctx, msg.ID, "failed", err.Error())
			return err
		}
	}

	if err := s.repo.UpdateImageResult(ctx, msg.ID, "done", processedPath, thumbPath); err != nil {
		return err
	}

	return nil
}

func saveJPEG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
}

// addSimpleWatermark добавляет простую светлую плашку в угол изображения.
func addSimpleWatermark(src image.Image) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(b)
	draw.Draw(dst, b, src, b.Min, draw.Src)

	wmWidth := 140
	wmHeight := 30

	startX := b.Max.X - wmWidth - 10
	startY := b.Max.Y - wmHeight - 10

	for y := startY; y < startY+wmHeight; y++ {
		for x := startX; x < startX+wmWidth; x++ {
			if x >= b.Min.X && x < b.Max.X && y >= b.Min.Y && y < b.Max.Y {
				dst.Set(x, y, image.White)
			}
		}
	}

	return dst
}

var _ = fmt.Sprintf
