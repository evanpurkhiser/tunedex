package main

type Track struct {
	ID          int64   `db:"rowid,omitempty"`
	FilePath    string  `db:"file_path"`
	FileHash    string  `db:"file_hash"`
	ArtworkHash string  `db:"artwork_hash"`
	Artist      string  `db:"artist"`
	Title       string  `db:"title"`
	Remixer     string  `db:"remixer"`
	Album       string  `db:"album"`
	Release     string  `db:"release"`
	Publisher   string  `db:"publisher"`
	DiscNumber  string  `db:"disc_number"`
	TrackNumber string  `db:"track_number"`
	Genre       string  `db:"genre"`
	Year        int     `db:"year"`
	BPM         float64 `db:"bpm"`
	Key         string  `db:"key"`
}
