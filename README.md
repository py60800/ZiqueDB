# ZiqueDB
ZiqueDB is a desktop application  intended  initially to manage my own collection of Irish tunes (Those I play or I want to learn). The main features are :
    •  Search engine
    • Score display
    • Tune annotation: (How you like it, how you play it, free comment...)
    • Midi playing
    • MP3 playing with speed and pitch adjustment
    • Tune editing (using MuseScore as an companion application)
    • Tune import in ABC format (typically from “thesession.org”)
    • sets, lists creation
    • ...
Important: As of now, ZiqueDB only runs on a Linux desktop.  A Windows version could be considered if there is a demand for it but some help would be needed.

Note that MP3 playing depends on your own collection MP3 files.

## Technical details
    • Go application using GTK3 
    • sqlite database with gorm (ORM)
    • depends on an external midi synthesizer such as  fluidsynth
    • Intensive use of MuseScore4 is used for tune editing  and format conversion (=>musicxml, svg)
    • Use of abc2xml (and therefore on python3) for ABC import
    • Use of rubberband library for speed and pitch adjustment when playing mp3
    • Tunes are stored in musescore format (.mscz) and derived files are created for different purposes (play midi and display)

## Prerequisites
    • GTK3
    • Fluidsynth
    • Rubberband library
    • Some sound system (ie : pipewire)
    • python3 (for ABC files import)
    • sqlite3 library
    • Musescore4 (that can be downloaded as a portable app from https://musescore.org/)
      
Debian packages: libgtk-3-common, pipewire, rubberband-cli, libsqlite3-0


