package sync

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path"

	"github.com/nfnt/resize"
)

// ArtworkIndexer implements the MetadataProcessor interface, providing the
// ability for artwork of tracks to be resized and stored on disk.
type ArtworkIndexer struct {
	// SavePath specifies where artwork should be stored.
	SavePath string

	// Size specifies the size of images to store.
	Size image.Point

	// Quality specifies the JPEG quality of the resize image.
	Quality int
}

func (a *ArtworkIndexer) ProcessTrack(track *indexedTrack) error {
	if len(track.artwork) == 0 {
		return nil
	}

	artPath := path.Join(a.SavePath, track.ArtworkHash+"-small.jpg")

	// Artwork already indexed
	if _, err := os.Stat(artPath); err == nil {
		return nil
	}

	art, _, err := image.Decode(bytes.NewReader(track.artwork))
	if err != nil {
		return err
	}

	artSmall := resize.Thumbnail(uint(a.Size.X), uint(a.Size.Y), art, resize.Lanczos3)

	artFile, err := os.Create(artPath)
	if err != nil {
		return err
	}

	defer artFile.Close()

	options := jpeg.Options{
		Quality: a.Quality,
	}

	return jpeg.Encode(artFile, artSmall, &options)
}
