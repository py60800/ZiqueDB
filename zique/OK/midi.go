package zique

import (
	"fmt"
	"log"
	"os"

	"sort"
	"sync"
	"ypmod/midi"
)

type MStart struct {
}
type MTimeEv struct {
	MLength  int
	Beats    int
	BeatType int
}

func (m MTimeEv) String() string {
	return fmt.Sprintf("TimeEvt: %v %v/%v", m.MLength, m.Beats, m.BeatType)
}

type PNoteOn struct {
	NoteNumber int
	Velocity   int
	Channel    int
}

type PNoteOff struct {
	NoteNumber int
	Channel    int
}
type PTickOn struct {
	NoteNumber int
	Velocity   int
}
type PTickOff struct {
	NoteNumber int
}
type PChordOn struct {
	p        *Player
	chordOff *PChordOff
	Velocity int
}
type PProgramChange struct {
	Patch int
}
type PTempoChange struct {
	Value int
}

func (t PTempoChange) String() string {
	return fmt.Sprintf("Tempo:", t.Value)
}
func (c PChordOn) String() string {
	return fmt.Sprintf("ChordOn (%v)", c.Velocity)
}

type PChordOff struct {
	Notes []int
}

func (c *PChordOff) String() string {
	return fmt.Sprintf("ChordOff")
}

type PEvent interface {
	GetMidiEvent() *midi.Event
	GetRtMidiEvent() []byte
	String() string
}

type SeqEvent struct {
	Tick  uint32
	Event PEvent
}
type SeqChord struct {
	Tick  uint32
	Chord EChord
}
type EChord struct {
	Key  int
	Mode string
}
type MidiFile struct {
	File    *os.File
	Encoder *midi.Encoder
	Track   *midi.Track
}

func (p *Player) KeyToInt(note string, octave int, alter int) int {
	n := midi.KeyInt(note, octave)
	n += alter
	return n
}

var DefaultChannel int = 0
var ChordChannel int = 2
var DrumChannel int = 9

func NewMidiFile(fileName string, Divisions int) MidiFile {
	var m MidiFile
	w, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	m.File = w

	m.Encoder = midi.NewEncoder(w, midi.SingleTrack, uint16(MasterDivisions))
	m.Track = m.Encoder.NewTrack()
	m.Track.SetName("The Track")
	m.Track.Add(0.0, midi.TempoEvent(float64(Tempo)))
	m.Track.Add(0.0, midi.ProgramChange(DefaultChannel, 0, MainInstrument))
	m.Track.Add(0.0, midi.ProgramChange(ChordChannel, 0, ChordInstrument))

	return m
}
func (m *MidiFile) Finish() {
	if err := m.Encoder.Write(); err != nil {
		log.Fatal(err)
	}
	m.File.Close()
}

func (t PTempoChange) GetMidiEvent() *midi.Event {
	Tempo = t.Value
	return midi.TempoEvent(float64(Tempo))
}

func (n PNoteOn) GetMidiEvent() *midi.Event {
	return midi.NoteOn(n.Channel, n.NoteNumber, n.Velocity)
}
func (n PNoteOn) String() string {
	return fmt.Sprintf("NoteOn %v %v", n.NoteNumber, n.Velocity)
}
func (n PTickOn) String() string {
	return fmt.Sprintf("--TickOn %v %v", n.NoteNumber, n.Velocity)
}
func (n PNoteOff) GetMidiEvent() *midi.Event {
	return midi.NoteOff(n.Channel, n.NoteNumber)
}
func (n PNoteOff) String() string {
	return fmt.Sprintf("NoteOff %v", n.NoteNumber)
}
func (n PTickOff) String() string {
	return fmt.Sprintf("--TickOff %v", n.NoteNumber)
}
func (n PTickOn) GetMidiEvent() *midi.Event {
	return midi.NoteOn(DrumChannel, n.NoteNumber, DrumVelocity)
}

func (n PTickOff) GetMidiEvent() *midi.Event {
	return midi.NoteOff(DrumChannel, n.NoteNumber)
}
func (n MStart) GetMidiEvent() *midi.Event {
	return nil
}
func (n MTimeEv) GetMidiEvent() *midi.Event {
	//p.ComputeSwingFactor(n.MLength, n.Beats, n.BeatType)
	return nil
}
func (c PChordOn) GetMidiEvent() *midi.Event {
	return nil
}
func (c *PChordOff) GetMidiEvent() *midi.Event {
	return nil
}
func (p PProgramChange) GetMidiEvent() *midi.Event {
	panic("Unimplemented")
	return nil
}
func (p PProgramChange) String() string {
	return fmt.Sprintf("PChange: %d", p.Patch)
}
func (m MStart) String() string {
	return fmt.Sprintf("MeasureStart")
}

type DummyEvent struct {
}

func (p DummyEvent) GetMidiEvent() *midi.Event {
	return nil
}
func (p DummyEvent) GetRtMidiEvent() []byte {
	return []byte{}
}
func (p DummyEvent) String() string {
	return "Dummy Event"
}

func (p *Player) ToMidi(fileName string) {
	mFile := NewMidiFile(fileName, p.XmlDivisions)
	sort.Slice(p.Sequence, func(i int, j int) bool { return p.Sequence[i].Tick < p.Sequence[j].Tick })

	Clock := 0

	for _, el := range p.Sequence {
		deltaT := int(el.Tick) - Clock
		evt := el.Event.GetMidiEvent()
		if evt != nil {
			mFile.Track.AddAfterDelta(uint32(deltaT), el.Event.GetMidiEvent())
			Clock = int(el.Tick)
		}
	}
	mFile.Finish()
}
func MidiRecord(pl *Player, mchan chan SeqEvent, ctrlChan chan int, wg *sync.WaitGroup) {
	for {
		select {
		case evt := <-mchan:
			pl.Sequence = append(pl.Sequence, evt)
			//			fmt.Println("Midi:", evt)

		case cmd := <-ctrlChan:
			switch cmd {
			case 0:
				fmt.Println("Midi: Finish")
				wg.Done()
				return
			case 1:
			}
		}

	}

}

// ****************************************************
