package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"upper.io/db.v3"
)

// Regex to split the artists names in the artist column
var artistSplitRegex = regexp.MustCompile(`(,| [Vv][Ss]\.?| &| Ft\.) `)

type ArtistPart struct {
	Value    string
	IsArtist bool
}

type FieldsService struct {
	// TrackCollection is the database upper.io db.Collection implementation
	// that may be queried on.
	TrackCollection db.Collection

	// cachedFields contain the cached list of various fields for the entire
	// collection database.
	cachedFields map[string][]string

	artistParts     map[string][]ArtistPart
	titleRemixParts map[string][]ArtistPart
}

func (s *FieldsService) RebuildFieldCache() error {
	var tracks []Track

	if err := s.TrackCollection.Find().All(&tracks); err != nil {
		return err
	}

	dedupeFields := map[string]map[string]bool{}
	cachedFields := map[string][]string{}

	artistParts := map[string][]ArtistPart{}

	for _, track := range tracks {
		fieldsToCache := map[string]string{
			"title":     track.Title,
			"album":     track.Album,
			"release":   track.Release,
			"publisher": track.Publisher,
			"genre":     track.Genre,
			"key":       track.Key,
		}

		// Cache simple string fields
		for field, val := range fieldsToCache {
			if _, ok := dedupeFields[field]; !ok {
				dedupeFields[field] = map[string]bool{}
				cachedFields[field] = []string{}
			}

			if val == "" || dedupeFields[field][val] {
				continue
			}

			dedupeFields[field][val] = true

			cachedFields[field] = append(cachedFields[field], val)
		}

		// Cache individual artist names and map full artist fields to the
		// ArtistPart list
		individualArtists, parts := splitArtists(track.Artist)

		cachedFields["artist"] = individualArtists
		artistParts[track.Artist] = parts

		// Cache individual remix artists into the artists map and

		//

	}

	s.cachedFields = cachedFields
	s.artistParts = artistParts

	derp, _ := json.Marshal(s.artistParts)

	fmt.Println(string(derp))

	return nil
}

// constructCachedFields extracts the simple fields from
func constructCachedFields(tracks []Track) map[string][]string {

}

// splitArtists accepts a string of artist that may contain multiple artists
// separated by the artistSplitRegex and will split them into individual
// artists and artist parts.
//
// The []ArtistPart will contain all text used to construct the full artist
// string, with individual artists marked.
func splitArtists(artist string) ([]string, []ArtistPart) {
	artists := artistSplitRegex.Split(artist, -1)
	separator := artistSplitRegex.FindAllString(artist, -1)

	take := map[int][]string{
		0: artists,
		1: separator,
	}

	var artistParts []ArtistPart

	// Join the separators back into the list, including information as to
	// weather the part is an artist or not. NOTE that we make the assumption
	// here that the first part will *always* be an artist.
	for i := 0; i < len(artists)+len(separator); i++ {
		artistParts = append(artistParts, ArtistPart{
			Value:    take[i%2][i/2],
			IsArtist: i%2 == 0,
		})
	}

	return artists, artistParts

}
