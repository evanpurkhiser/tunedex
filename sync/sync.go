package sync

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rjeczalik/notify"
	"upper.io/db.v3"

	"go.evanpurkhiser.com/tunedex/data"
	"go.evanpurkhiser.com/tunedex/metadata"
)

var defaultTrackFiletypes = []string{"aif", "mp3"}

type indexedTrack struct {
	data.Track
	artwork  []byte
	realPath string
}

// A MetadataProcessor is an interface that defines a module that can be added
// to the MetadataIndexer to perform additional processing when adding,
// removing, or changing a track in the database.
type MetadataProcessor interface {
	ProcessTrack(*indexedTrack) error
}

// MetadataIndexer is a service object that handles watching a directory
// containing a collection of music for new and changed tracks and will index
// them into the provided database collection.
type MetadataIndexer struct {
	// CollectionPath specifies the location of the music collection on disk to
	// keep in sync with the database.
	CollectionPath string

	// TrackCollection is the database upper.io db.Collection implementation
	// that may be queried on.
	TrackCollection db.Collection

	// TrackFiletypes specifies the types of files supported in the collection.
	// This defaults to aif and mp3.
	TrackFiletypes []string

	// Processors is a list of MetadataProcessors that will be executed when
	// indexing a track.
	Processors []MetadataProcessor
}

// isValidFiletype checks that the provided path is part of the valid track
// filetypes list.
func (i *MetadataIndexer) isValidFiletype(path string) bool {
	types := i.TrackFiletypes

	if types == nil {
		types = defaultTrackFiletypes
	}

	for _, extension := range types {
		if filepath.Ext(path)[1:] == extension {
			return true
		}
	}

	return false
}

// getAllFiles finds all media files in the CollectionPath
func (i *MetadataIndexer) getAllFiles() ([]string, error) {
	paths := []string{}

	walker := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() || err != nil {
			return nil
		}

		if !i.isValidFiletype(path) {
			return nil
		}

		paths = append(paths, path)

		return nil
	}

	err := filepath.Walk(i.CollectionPath, walker)

	return paths, err
}

// buildTrack constructs the Track object given a path to the track.
func (i *MetadataIndexer) buildTrack(path string) (*indexedTrack, error) {
	metadata, err := metadata.ForTrack(path)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	trackSum := hash.Sum(nil)
	artworkSum := md5.Sum(metadata.Artwork)

	year, _ := strconv.Atoi(metadata.Year)

	trackData := data.Track{
		FilePath:    path,
		FileHash:    fmt.Sprintf("%x", trackSum),
		ArtworkHash: fmt.Sprintf("%x", artworkSum),
		Artist:      metadata.Artist,
		Title:       metadata.Title,
		Album:       metadata.Album,
		Remixer:     metadata.Remixer,
		Publisher:   metadata.Publisher,
		Release:     metadata.Release,
		TrackNumber: metadata.Track,
		DiscNumber:  metadata.Disc,
		Genre:       metadata.Genre,
		Key:         metadata.Key,
		Year:        year,
	}

	track := indexedTrack{
		Track:    trackData,
		artwork:  metadata.Artwork,
		realPath: path,
	}

	return &track, nil
}

func (i *MetadataIndexer) trackAdded(track *indexedTrack) error {
	count, err := i.TrackCollection.Find("file_hash =", track.FileHash).Count()
	if err != nil {
		return err
	}

	if count != 0 {
		return fmt.Errorf("Track already exists in database")
	}

	if _, err := i.TrackCollection.Insert(track); err != nil {
		return err
	}

	for _, processor := range i.Processors {
		processor.ProcessTrack(track)
	}

	return nil
}

func (i *MetadataIndexer) trackModified(track *indexedTrack) error {
	err := i.TrackCollection.Find("file_path =", track.FilePath).Update(track)

	for _, processor := range i.Processors {
		processor.ProcessTrack(track)
	}

	return err
}

func (i *MetadataIndexer) trackMoved(track *indexedTrack) error {
	err := i.TrackCollection.Find("file_hash =", track.FileHash).Update(track)

	return err
}

func (i *MetadataIndexer) WatchCollection() error {
	events := make(chan notify.EventInfo, 1)

	// We specifically do *not* handle removal of files, this is to save us
	// from losing metadata in an accidental delete.
	watchEvents := []notify.Event{
		notify.InCreate,
		notify.InCloseWrite,
		notify.InMovedTo,
	}

	// The '...' syntax is used in the notify library for recursive watching
	path := filepath.Join(i.CollectionPath, "...")

	if err := notify.Watch(path, events, watchEvents...); err != nil {
		return err
	}

	handlers := map[notify.Event]func(*data.Track) error{
		notify.InCreate:     i.trackAdded,
		notify.InMovedTo:    i.trackMoved,
		notify.InCloseWrite: i.trackModified,
	}

	for eventInfo := range events {
		path := eventInfo.Path()

		if !i.isValidFiletype(path) {
			continue
		}

		fmt.Println(eventInfo)

		track, err := i.buildTrack(path)
		if err != nil {
			log.Printf("Failed to construct track to index: %q", err)
			continue
		}

		err = handlers[eventInfo.Event()](track)
		if err != nil {
			log.Printf("Failed to index track: %q", err)
			continue
		}
	}

	return nil
}

// IndexAll
func (i *MetadataIndexer) IndexAll() error {
	collection, err := i.getAllFiles()
	if err != nil {
		return err
	}

	for _, path := range collection {
		track, err := i.buildTrack(path)
		if err != nil {
			log.Printf("Failed to construct track to index: %q", err)
			continue
		}

		err = i.trackAdded(track)
		if err != nil {
			log.Printf("Failed to index track: %q", err)
			continue
		}
	}

	return nil
}
