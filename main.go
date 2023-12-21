package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/vilmibm/drift/game"
)

const (
	minHeight    = 10
	minWidth     = 12
	animInterval = time.Millisecond * 300
)

type flake struct {
	game.GameObject
	speed int
	HP    int
}

func (f *flake) Update() {
	if f.HP <= 0 {
		// TODO could make some water
		f.Game.Destroy(f)
		return
	}
	f.Y += f.speed
	if f.Y > f.Game.MaxHeight {
		f.Y = f.Game.MaxHeight
	}

	if f.Y == f.Game.MaxHeight {
		f.HP--
		// TODO drain color
	}
}

func newFlake(g *game.Game, char rune) *flake {
	x := rand.Intn(g.MaxWidth)
	y := rand.Intn(5)
	colorOffset := int32(rand.Intn(50))
	so := g.Style.Foreground(
		tcell.NewRGBColor(255-colorOffset, 255-colorOffset, 255-colorOffset))
	speed := rand.Intn(5)
	hpOffset := rand.Intn(5)
	return &flake{
		GameObject: game.GameObject{
			Game:          g,
			X:             x,
			Y:             y,
			W:             1,
			H:             1,
			Sprite:        string(char),
			StyleOverride: &so,
		},
		speed: speed,
		HP:    10 + hpOffset,
	}
}

type wind struct {
	game.GameObject
	direction int
	speed     int
}

func (s *wind) Update() {
	s.X = s.X + (s.direction * s.speed)
}

func newWind(g *game.Game) *wind {
	x := 0
	y := 5     // TODO randomize
	width := 5 // TODO randomize
	speed := 5 // TODO randomize
	dir := 1   // TODO randomize

	spr := ""
	for w := 0; w < width; w++ {
		spr += " "
	}

	return &wind{
		GameObject: game.GameObject{
			X: x, Y: y,
			W:      width,
			H:      1,
			Sprite: spr,
			Game:   g,
		},
		direction: dir,
		speed:     speed,
	}
}

func _main(lines []string) (err error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}

	if err = s.Init(); err != nil {
		return err
	}

	defer s.Fini()

	defStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	s.SetStyle(defStyle)

	quit := make(chan struct{})
	gust := make(chan struct{})
	go inputLoop(s, gust, quit)()

	w, h := s.Size()
	if w < minWidth || h < minHeight {
		return errors.New("terminal is too small i'm sorry")
	}

	gg := &game.Game{
		Screen:    s,
		Style:     defStyle,
		MaxWidth:  w,
		MaxHeight: h,
	}

	rand.Seed(time.Now().Unix())

	dogust := func() {
		// TODO generate a bunch of wind
	}

	var lineIX int
	var quitting bool
	for {
		select {
		case <-quit:
			quitting = true
		case <-gust:
			dogust()
		case <-time.After(animInterval):
		}

		if quitting {
			break
		}

		chance := rand.Intn(100)
		if chance < 50 {
			rline := []rune(lines[lineIX])

			for ix := 0; ix < len(rline); ix++ {
				gg.AddDrawable(newFlake(gg, rline[ix]))
			}

			lineIX++
			if lineIX >= len(lines) {
				lineIX = 0
			}
		}

		// TODO wind generation

		s.Clear()
		gg.Update()
		gg.Draw()
		s.Show()
	}

	return nil
}

func inputLoop(s tcell.Screen, gust chan struct{}, quit chan struct{}) func() {
	return func() {
		for {
			s.Show()

			ev := s.PollEvent()

			switch ev := ev.(type) {
			case *tcell.EventResize:
				s.Sync()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEnter {
					gust <- struct{}{}
				}
				if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
					close(quit)
				}
			}
		}
	}
}

func main() {
	lines := []string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if scanner.Err() != nil {
		os.Exit(2)
	}

	err := _main(lines)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
