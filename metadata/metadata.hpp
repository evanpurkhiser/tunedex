#ifndef METADATA_HPP
#define METADATA_HPP

#ifdef __cplusplus
extern "C" {
#endif

typedef struct {
    const char* artist;
    const char* title;
    const char* album;
    const char* remixer;
    const char* publisher;
    const char* comment;
    const char* key;
    const char* bpm;
    const char* year;
    const char* track_number;
    const char* disc_number;
    const char* genre;
    const char* artwork;
    int         art_size;
} track;

track* metadata(const char* path);

#ifdef __cplusplus
}
#endif

#endif
