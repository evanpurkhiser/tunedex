package main

import (
	"upper.io/db.v3/sqlite"
)

const CollectionPath = "/mnt/documents/multimedia/djing/tracks/"

func main() {
	dbConfig := sqlite.ConnectionURL{
		Database: "database/database.db3",
	}

	sess, err := sqlite.Open(dbConfig)
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	tracksColl := sess.Collection("tracks")

	//	indexer := MetadataIndexer{
	//		CollectionPath:  CollectionPath,
	//		TrackCollection: tracksColl,
	//	}

	listing := FieldsService{
		TrackCollection: tracksColl,
	}

	listing.RebuildFieldCache()

}
