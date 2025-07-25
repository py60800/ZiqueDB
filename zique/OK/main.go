// goxml0 project main.go
package zique

import (
	"bufio"
	//	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"
)

var RtMidiChan chan SeqEvent

/*
	type DrumBeat struct {
		Instrument int
		Velocity   int
		Duration   float64
	}
*/
var (
	Tempo           = 120
	MainInstrument  = 0
	Velocity        = 100
	ChordInstrument = 0
	ChordVelocity   = 0
	DrumVelocity    = 0
	//	SwingPattern     = []float64{}
	//VelocityPattern = []float64{}
	MelodyOff = false

// DrumPattern     = []DrumBeat{}
)

var ChordPattern []ChordStroke

const MasterDivisions = 1200

type SetElem struct {
	File  string
	Count int
	Tempo int
}
type MusicSet []SetElem

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ln := scanner.Text()
		if !strings.HasPrefix(ln, "#") {
			lines = append(lines, scanner.Text())
		}
	}
	return lines, scanner.Err()
}

var RtCtrlChan chan int

type ZiquePlayer struct {
	ZiqueCtrl       chan interface{}
	DrumPattern     string
	VelocityPattern string
	SwingPattern    string
	player          *Player
	FeedBack        chan string
	TickBack        chan Tick
	RtFeedBack      chan time.Duration
}

func ZiquePlayerNew() *ZiquePlayer {
	z := ZiquePlayer{
		ZiqueCtrl: make(chan interface{}, 10),
		FeedBack:  make(chan string, 2),
		TickBack:  make(chan Tick, 2),
	}

	return &z
}

func (z *ZiquePlayer) Init() {

	InitPattern()
	RtMidiChan = make(chan SeqEvent, 3)
	RtCtrlChan = make(chan int, 3)
	z.RtFeedBack = make(chan time.Duration, 2)

	go RtPlay(RtMidiChan, RtCtrlChan, z.RtFeedBack)

	go z.mainLoop(z.ZiqueCtrl)
}

func (z *ZiquePlayer) SetTempo(tempo int) {
	RtMidiChan <- SeqEvent{0, PTempoChange{tempo}}
}
func (z *ZiquePlayer) SetPatch(patch string) {
	p, ok := RPatch[patch]
	if ok {
		RtMidiChan <- SeqEvent{0, PProgramChange{p - 1}}
	}

}
func (z *ZiquePlayer) SetMainVolume(v int) {
	Velocity = v
}
func (z *ZiquePlayer) SetDrumVolume(v int) {
	DrumVelocity = v
}
func (z *ZiquePlayer) SetDrumPattern(dp string) {
	z.DrumPattern = dp
	z.player.PlayCtrl <- CPlayCtrl{CDRUM, dp, 1.0}

}
func (z *ZiquePlayer) SetSwingPattern(dp string) {
	z.SwingPattern = dp
	z.player.PlayCtrl <- CPlayCtrl{CSWING, dp, 1.0}

}
func (z *ZiquePlayer) ALterSwingPattern(coeff float64) {
	z.player.PlayCtrl <- CPlayCtrl{CSWING, z.SwingPattern, coeff}

}

func (z *ZiquePlayer) SetVelocityPattern(dp string) {
	z.VelocityPattern = dp
	z.player.PlayCtrl <- CPlayCtrl{CVELOCITY, dp, 1.0}

}
func (z *ZiquePlayer) AlterVelocityPattern(coeff float64) {
	z.player.PlayCtrl <- CPlayCtrl{CVELOCITY, z.VelocityPattern, coeff}
}

func (z *ZiquePlayer) Kill() {
	z.ZiqueCtrl <- 0
}
func (z *ZiquePlayer) Play(tune string) {
	set := []SetElem{SetElem{File: tune, Count: 0}}
	z.PlaySet(set)
}
func (z *ZiquePlayer) PlaySet(tunes MusicSet) {
	RtCtrlChan <- 3 // Resume Pause
	fmt.Println("ZPlay:", tunes)
	z.ZiqueCtrl <- tunes
}
func (z *ZiquePlayer) Pause() {
	RtCtrlChan <- 1
}
func (z *ZiquePlayer) Stop() {
	fmt.Print("Stop")
	z.ZiqueCtrl <- 1
	fmt.Println("..ping!")
}

func (z *ZiquePlayer) mainLoop(Cmd chan interface{}) {
	var wg sync.WaitGroup
	pl := MakePlayer("Dummy")
	z.player = &pl
	var set MusicSet
	Playing := false

	PlayingThread := func() {
		fmt.Println("Start playing")
		wg.Add(1)
		defer func() {
			Playing = false
			wg.Done()
			fmt.Println("Exit playing")
		}()
		fmt.Println("Play:", set)
		for iset, t := range set {
			fmt.Println("Parse:", t.File)
			select {
			case z.FeedBack <- t.File:
			default:
			}
			partition, _ := Parse(t.File)
			count := t.Count
			barCount := -1
			signalBar := -1
			for Playing {
				barCount = pl.PlayTune(partition.Part[0], signalBar, func() {
					if iset != len(set)-1 {
						select {
						case z.FeedBack <- set[iset+1].File:
						default:
						}

					}
				})
				if count > 0 { // count == 0 => infinite
					count--
					if count == 1 && barCount > 0 {
						signalBar = barCount - 2
						barCount = -1
					}
					if count == 0 {
						break
					}
				}
			}
			if !Playing {
				break
			}
		}

	}

	for {
		select {
		case TickTime := <-z.RtFeedBack:

			select {
			case z.TickBack <- Tick{Beats: z.player.Beats, BeatType: z.player.BeatType,
				XmlDivisions: z.player.XmlDivisions,
				TickTime:     TickTime}:
			default:
			}

		case cmd := <-Cmd:
			fmt.Println("Player: Cmd received:", cmd)
			switch v := cmd.(type) {
			case MusicSet:
				if Playing {
					pl.PlayCtrl <- CPlayCtrl{CSTOP, "", 0.0}
					Playing = false
					fmt.Println("Player: Wg wait")
					wg.Wait()
				}

				set = v
				Playing = true
				go PlayingThread()

			case int:
				switch v {
				case 0:
					pl.PlayCtrl <- CPlayCtrl{CSTOP, "", 0.0}
					if Playing {
						Playing = false
						fmt.Println("Player: Wg wait")
						wg.Wait()
					}
					fmt.Println("Player: Exit playing thread")
					return
				case 1:
					if Playing {
						pl.PlayCtrl <- CPlayCtrl{CSTOP, "", 0.0}
						Playing = false
						fmt.Println("Player: Wg wait")
						wg.Wait()
					}
					fmt.Println("Player: Stopped")

				}
			}
		}
	}
}
