#include "metadata.hpp"

#include <taglib/fileref.h>
#include <taglib/id3v2tag.h>
#include <taglib/mpegfile.h>
#include <taglib/aifffile.h>
#include <taglib/attachedpictureframe.h>
#include <iostream>

typedef std::pair<TagLib::File*, TagLib::ID3v2::Tag*> fileTags;

fileTags get_tags(const char* path)
{
    const auto path_str = std::string(path);
    const int ext_pos = path_str.rfind(".");

    if (ext_pos == -1)
        return {0, 0};

    auto ext = path_str.substr(ext_pos + 1);

    // AIFF file, access the ID3 tags directly using AIFF::File::tag
    if (ext == "aif" || ext == "aiff")
    {
        auto file = new TagLib::RIFF::AIFF::File(path);
        return fileTags{file, file->tag()};
    }

    // MP3 file, access the ID3 tags using MPEG::File::ID3v2Tag
    if (ext == "mp3")
    {
        auto file = new TagLib::MPEG::File(path);
        return fileTags{file, file->ID3v2Tag()};
    }

    return {0, 0};
}

/**
 * Retrieve tag information given a file path. Currently only MP3 and AIFF
 * files are supported.
 */
track* metadata(const char* path)
{
    auto file_tags = get_tags(path);

    auto file = file_tags.first;
    auto tags = file_tags.second;

    if (!file || !tags) return 0;

    auto frames = tags->frameListMap();
    auto metadata = new track();

    std::map<std::string, TagLib::ID3v2::FrameList> track_frames = {
        {"artist",       frames["TPE1"]},
        {"title",        frames["TIT2"]},
        {"album",        frames["TALB"]},
        {"remixer",      frames["TPE4"]},
        {"publisher",    frames["TPUB"]},
        {"comment",      frames["COMM"]},
        {"key",          frames["TKEY"]},
        {"bpm",          frames["TBPM"]},
        {"year",         frames["TDRC"]},
        {"track_number", frames["TRCK"]},
        {"disc_number",  frames["TPOS"]},
        {"genre",        frames["TCON"]},
    };

    std::map<std::string, char*> strings;

    for (auto const &kv : track_frames)
    {
        if (kv.second.isEmpty()) continue;

        // Get the frame value as a utf8 std::string
        auto str = kv.second.front()->toString().to8Bit(true);

        // Copy the value into memory
        char* str_copy = new char[str.size() + 1];
        std::copy(str.begin(), str.end(), str_copy);
        str_copy[str.size()] = '\0';

        strings.emplace(kv.first, str_copy);
    }

    auto art_frames = frames["APIC"];

    if (!art_frames.isEmpty())
    {
        auto frame   = (TagLib::ID3v2::AttachedPictureFrame*) art_frames.front();
        auto artwork = frame->picture();

        auto art_data = new char[artwork.size()];
        std::copy(artwork.begin(), artwork.end(), art_data);

        metadata->artwork = art_data;
        metadata->art_size = artwork.size();
    }

    metadata->artist       = strings["artist"];
    metadata->title        = strings["title"];
    metadata->album        = strings["album"];
    metadata->remixer      = strings["remixer"];
    metadata->publisher    = strings["publisher"];
    metadata->comment      = strings["comment"];
    metadata->key          = strings["key"];
    metadata->bpm          = strings["bpm"];
    metadata->year         = strings["year"];
    metadata->track_number = strings["track_number"];
    metadata->disc_number  = strings["disc_number"];
    metadata->genre        = strings["genre"];

    delete file;
    return metadata;
}
