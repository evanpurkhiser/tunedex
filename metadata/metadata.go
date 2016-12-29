package metadata

import (
	// #cgo LDFLAGS: -ltag
	// #include "metadata.hpp"
	"C"
	"fmt"
	"unsafe"
)

// Track contains the tracks metadata.
type Track struct {
	Artist    string `json:"artist"`
	Title     string `json:"title"`
	Album     string `json:"album"`
	Remixer   string `json:"remixer"`
	Publisher string `json:"publisher"`
	Release   string `json:"release"`
	Key       string `json:"key"`
	BPM       string `json:"bpm"`
	Year      string `json:"year"`
	Track     string `json:"track"`
	Disc      string `json:"disc"`
	Genre     string `json:"genre"`
	Artwork   []byte `json:"-"`
}

// ForTrack retrieves track metadata given file path.
func ForTrack(path string) (*Track, error) {
	track := C.metadata(C.CString(path))

	if track == nil {
		return nil, fmt.Errorf("Failed to read metadata for file %s", path)
	}

	metadata := Track{
		Artist:    C.GoString(track.artist),
		Title:     C.GoString(track.title),
		Album:     C.GoString(track.album),
		Remixer:   C.GoString(track.remixer),
		Publisher: C.GoString(track.publisher),
		Release:   C.GoString(track.comment),
		Key:       C.GoString(track.key),
		BPM:       C.GoString(track.bpm),
		Year:      C.GoString(track.year),
		Track:     C.GoString(track.track_number),
		Disc:      C.GoString(track.disc_number),
		Genre:     C.GoString(track.genre),
		Artwork:   C.GoBytes(unsafe.Pointer(track.artwork), track.art_size),
	}

	return &metadata, nil
}
