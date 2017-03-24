package download

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
	"upper.io/db.v3"

	"go.evanpurkhiser.com/tunedex/data"
)

// An ArchiveJob includes details about an archiving job.
type ArchiveJob struct {
	Name   string
	Tracks []*data.Track

	// Completed is a channel that will be written into when the job has
	// successfully finished archving all files and is ready to be requested.
	Completed chan bool

	// Error is a channel that will be written into if there are any problems
	// while creating the archive.
	Error chan error

	// Progress is a channel that will be written into when a file is starting
	// to be added into the in progress archive.
	Progress chan *data.Track

	// Cancel may be written into to stop the archiving job.
	Cancel chan bool

	// archiveFile is the temporary file that the archive is written to.
	archiveFile *os.File

	// requested should be written into after the archive is completed and it
	// has been requested. This will stop the archive from being removed before
	// anything can be done with the file.
	requested chan bool
	completed bool
}

// Archiver is a service object which when given a list of tracks, will archive
// them into a ZIP file. This zip file may then be requested after all tracks
// have been fully archived.
//
// If a file is not requested after a given interval, it will be removed and
// can no longer be requested.
type Archiver struct {
	// AllowedRequestInterval specifies how long to store the archive between
	// when it has been created and when it is requested.
	AllowedRequestInterval time.Duration

	// CollectionPath specifies the location of the music collection on disk.
	CollectionPath string

	// TrackCollection is the database upper.io db.Collection implementation
	// that may be queried on.
	TrackCollection db.Collection

	// inProgress tracks current in progress of in flight archive jobs.
	inProgress map[string]*ArchiveJob
}

// Create constructs an ArchiveJob and being the background archiving process.
func (a *Archiver) Create(tracks []*data.Track) (*ArchiveJob, error) {
	trackCount := len(tracks)

	if trackCount == 0 {
		return nil, fmt.Errorf("No tracks provided to archive")
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	name := uuid.String()

	file, err := ioutil.TempFile("", name)
	if err != nil {
		return nil, err
	}

	job := &ArchiveJob{
		Name:        name,
		Tracks:      tracks,
		Completed:   make(chan bool, 1),
		Cancel:      make(chan bool, 1),
		Error:       make(chan error, 1),
		Progress:    make(chan *data.Track, trackCount),
		requested:   make(chan bool, 1),
		archiveFile: file,
	}

	if a.inProgress == nil {
		a.inProgress = map[string]*ArchiveJob{}
	}

	a.inProgress[job.Name] = job

	go a.archive(job)

	return job, nil
}

// Request returns the os.File containing the completed archive. If the
// provided archive job is incomplete or no longer exists no file will be
// returned.
//
// It is the responsability of the caller to close and remove the file.
func (a *Archiver) Request(name string) *os.File {
	job, ok := a.inProgress[name]
	if !ok || !job.completed {
		return nil
	}

	job.requested <- true

	a.cleanupJob(job)

	return job.archiveFile
}

func (a *Archiver) archive(job *ArchiveJob) {
	zipFile := zip.NewWriter(job.archiveFile)

	// Copy the file list into the zip
	for _, track := range job.Tracks {
		partPath := strings.TrimPrefix(track.FilePath, a.CollectionPath)
		zipPart, err := zipFile.Create(partPath)

		if err != nil {
			job.Error <- err
			return
		}

		file, err := os.Open(track.FilePath)
		if err != nil {
			job.Error <- err
			return
		}

		io.Copy(zipPart, file)
		file.Close()

		job.Progress <- track
	}

	if err := zipFile.Close(); err != nil {
		job.Error <- err
		return
	}

	// The archive has been created, start the clock on removing it if it's
	// never requested
	go a.waitToRemove(job)

	job.completed = true
	job.Completed <- true
}

func (a *Archiver) waitToRemove(job *ArchiveJob) {
	timer := time.NewTimer(a.AllowedRequestInterval)

	select {
	case <-job.requested:
		return
	case <-timer.C:
	}

	job.archiveFile.Close()
	os.Remove(job.archiveFile.Name())

	a.cleanupJob(job)
}

func (a *Archiver) cleanupJob(job *ArchiveJob) {
	delete(a.inProgress, job.Name)

	close(job.Cancel)
	close(job.Completed)
	close(job.Error)
	close(job.Progress)
	close(job.requested)
}
