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
	hp    int
	color int32
}

func (f *flake) Update() {
	if f.hp <= 0 {
		f.Game.Destroy(f)
		return
	}
	ground := f.Game.MaxHeight - 1
	next := f.Y + f.speed
	flakeFilter := func(d game.Drawable) bool {
		if _, ok := d.(*flake); !ok {
			return false
		}
		p := d.Pos()
		return p.Y == next && p.X == f.X
	}
	for len(f.Game.FilterGameObjects(flakeFilter)) > 0 && next > f.Y {
		next--
	}

	f.Y = next
	if f.Y > ground {
		f.Y = ground
	}

	windFilter := func(d game.Drawable) bool {
		if _, ok := d.(*wind); !ok {
			return false
		}
		p := d.Pos()
		s := d.Size()
		if f.Y != p.Y {
			return false
		}

		return f.X <= p.X && f.X > p.X-s.X
	}

	winds := f.Game.FilterGameObjects(windFilter)
	if len(winds) > 0 {
		w := winds[0]
		f.X += w.(*wind).Speed
	}

	if f.Y == ground {
		f.hp--
		f.color -= 5
		if f.color < 25 {
			f.color = 25
		}
		so := f.Game.Style.Foreground(tcell.NewRGBColor(f.color, f.color, f.color))
		f.StyleOverride = &so
	}
}

func newFlake(g *game.Game, x int, char rune) *flake {
	y := rand.Intn(2)
	color := 255 - int32(rand.Intn(100))
	so := g.Style.Foreground(
		tcell.NewRGBColor(color, color, color))
	//speed := rand.Intn(3) + 1
	speed := 1
	hpOffset := rand.Intn(25)
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
		color: color,
		speed: speed,
		hp:    100 + hpOffset,
	}
}

type wind struct {
	game.GameObject
	direction int
	Speed     int
}

func (s *wind) Update() {
	s.X = s.X + (s.direction * s.Speed)
}

func newWind(g *game.Game) *wind {
	x := 0
	y := rand.Intn(g.MaxHeight - 4)
	width := rand.Intn(20) + 1
	speed := rand.Intn(10) + 3
	dir := 1 // TODO can spawn and go other way

	return &wind{
		GameObject: game.GameObject{
			X: x, Y: y,
			W:         width,
			H:         1,
			Game:      g,
			Invisible: true,
		},
		direction: dir,
		Speed:     speed,
	}
}

func _main(lines []string) (err error) {
	// IDEA z plane; only snow in foreground piles up; gives sense of depth (did
	// 			it by accident with the color offset thing)
	// TODO wind mechanic
	// TODO gust mechanic
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
		for y := 0; y < gg.MaxHeight; y++ {
			gg.AddDrawable(newWind(gg))
		}
	}

	var lineIX int
	var quitting bool
	starting := true
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
		if chance < 10 || starting {
			rline := []rune(lines[lineIX])

			x := 0
			for ix := 0; ix < len(rline); ix++ {
				gap := rand.Intn(gg.MaxWidth/len(rline)) + 1
				if ix == 0 {
					gap = rand.Intn(2)
				}
				if gap > 5 {
					gap -= 5
				}
				x += gap
				gg.AddDrawable(newFlake(gg, x, rline[ix]))
			}

			lineIX++
			if lineIX >= len(lines) {
				lineIX = 0
			}
		}

		starting = false

		windChance := rand.Intn(100)
		if windChance < 5 {
			//gg.AddDrawable(newWind(gg))
			dogust()
		}

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
