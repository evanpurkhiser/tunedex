CREATE TABLE "tracks" (
	file_hash    CHAR(32) UNIQUE,
	artwork_hash CHAR(32) NULL,
	file_path    VARCHAR(255),
	artist       VARCHAR(255),
	title        VARCHAR(255),
	remixer      VARCHAR(255),
	album        VARCHAR(255),
	release      VARCHAR(255),
	publisher    VARCHAR(255),
	disc_number  VARCHAR(7),
	track_number VARCHAR(7),
	genre        VARCHAR(255),
	year         INT(4),
	bpm          DECIMAL(5, 2),
	key          VARCHAR(3)
);

CREATE INDEX tracks_file_hash_idx ON tracks (file_hash);
CREATE INDEX tracks_release_idx ON tracks (release);
CREATE INDEX tracks_publisher_idx ON tracks (publisher);
CREATE INDEX tracks_genre_idx ON tracks (genre);
CREATE INDEX tracks_year_idx ON tracks (year);
CREATE INDEX tracks_bpm_idx ON tracks (bpm);
CREATE INDEX tracks_key_idx ON tracks (key);
