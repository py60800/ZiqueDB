// rtmidi.go
package zique

import (
	"fmt"
	//	"sort"
	"time"

	rtmidi "gitlab.com/gomidi/midi/v2"

	"os/exec"

	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver

	"sync"
)

func (n PNoteOn) GetRtMidiEvent() []byte {
	return rtmidi.NoteOn(uint8(n.Channel), uint8(n.NoteNumber), uint8(n.Velocity))
}
func (n PNoteOff) GetRtMidiEvent() []byte {
	return rtmidi.NoteOff(uint8(n.Channel), uint8(n.NoteNumber))
}
func (n PTickOn) GetRtMidiEvent() []byte {
	return rtmidi.NoteOn(uint8(DrumChannel), uint8(n.NoteNumber), uint8(n.Velocity))
}

func (n PTickOff) GetRtMidiEvent() []byte {
	return rtmidi.NoteOff(uint8(DrumChannel), uint8(n.NoteNumber))
}
func (n MStart) GetRtMidiEvent() []byte {
	return []byte{}
}
func (n MTimeEv) GetRtMidiEvent() []byte {
	//p.ComputeSwingFactor(n.MLength, n.Beats, n.BeatType)
	return []byte{}
}
func (c PChordOn) GetRtMidiEvent() []byte {
	return []byte{}
}
func (c *PChordOff) GetRtMidiEvent() []byte {
	return []byte{}
}
func (p PProgramChange) GetRtMidiEvent() []byte {
	return rtmidi.ProgramChange(uint8(DefaultChannel), uint8(p.Patch))
}
func (p PTempoChange) GetRtMidiEvent() []byte {
	return []byte{}
}

var MidiMutex sync.Mutex

func GetMidiSink() func(rtmidi.Message) error {
	tlimit := time.Now().Add(5 * time.Second)
	synthLaunched := false
	for {
		if out, err := rtmidi.FindOutPort("Synth"); err == nil {
			fmt.Println("Got synth")
			send, err := rtmidi.SendTo(out)
			if err != nil {
				panic(err)
			}
			return send
		}
		if tlimit.Before(time.Now()) {
			panic("Failed to launch synth(a)")
		}
		if !synthLaunched {
			cmd := exec.Command("fluidsynth", "-a", "alsa", "-i", "-s")
			fmt.Println("Start synth")
			err := cmd.Start()
			if err != nil {
				panic(err)
			}
			synthLaunched = true
		}

		time.Sleep(500 * time.Millisecond)
	}
	panic("Synth Failure(b)")
}

func RtPlay(mchan chan SeqEvent, ctrlChan chan int, feedBack chan time.Duration) {
	defer rtmidi.CloseDriver()
	/*
		out, err := rtmidi.FindOutPort("Synth")
		if err != nil {
			fmt.Printf("can't find qsynth")
			return
		}
		rtmidi.Send
		send, err := rtmidi.SendTo(out)
	*/
	send := GetMidiSink()
	for _, m := range rtmidi.SilenceChannel(-1) {
		send(m)
	}
	silence := func() {
		for _, m := range rtmidi.SilenceChannel(-1) {
			send(m)
		}

	}
	send(rtmidi.ProgramChange(uint8(DefaultChannel), uint8(MainInstrument)))
	send(rtmidi.ProgramChange(uint8(ChordChannel), uint8(ChordInstrument)))

	TickTime := time.Minute / time.Duration(Tempo*MasterDivisions)

	clock := uint32(0)
	//	fmt.Println("RtPlay Start")
	lastEvent := SeqEvent{0, DummyEvent{}}
	for {
		select {
		case evt := <-mchan:
			//fmt.Println(evt)
			if evt.Tick == 0 {
				evt.Tick = clock
			}
			switch v := evt.Event.(type) {
			case PTempoChange:
				Tempo = v.Value
				TickTime = time.Minute / time.Duration(Tempo*MasterDivisions)
			case MStart:
				select {
				case feedBack <- TickTime:
				default:
				}

			}
			ntime := evt.Tick
			wait := time.Duration(ntime-clock) * TickTime
			if wait < 0 || wait > 2*time.Second {
				fmt.Printf("Sequence error Event:%v (PrevEvent %d:%v) ntime:%v clock%v ntime-nclock:%v clock-ntime:%v\n ",
					evt.Event.String(),
					lastEvent.Tick, lastEvent.Event.String(),
					ntime, clock, ntime-clock, clock-ntime)

			} else {
				time.Sleep(wait)
			}
			clock = ntime
			b := evt.Event.GetRtMidiEvent()
			if len(b) > 0 {
				send(b)
			}
			lastEvent = evt

		case cmd := <-ctrlChan:
			fmt.Println("Ctrl chan:", cmd)
			switch cmd {
			case 0:
				silence()

				return
			case 1: // Pause
				silence()
				cmd = <-ctrlChan
				if cmd == 0 {
					return
				}
			case 2: // Stop
				silence()
				return
			case 3: // resume pause

			}
		}

	}
}
