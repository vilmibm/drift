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
	minHeight     = 10
	minWidth      = 12
	animInterval  = time.Millisecond * 300
	farFlakeLayer = -1
	windLayer     = 0
	flakeLayer    = 1
)

type farFlake struct {
	game.GameObject
	speed int
	color int32
}

func newFarFlake(g *game.Game, x int, char rune) *farFlake {
	color := int32(rand.Intn(30) + 1)
	speed := rand.Intn(2) + 1
	return &farFlake{
		GameObject: game.GameObject{
			Game:   g,
			X:      x,
			Y:      0,
			W:      1,
			H:      1,
			Sprite: string(char),
			Layer:  farFlakeLayer,
		},
		speed: speed,
		color: color,
	}
}

func (ff *farFlake) Update() {
	ff.color += int32(rand.Intn(20) - 10) // -10 to 10
	so := ff.Game.Style.Foreground(tcell.NewRGBColor(ff.color, ff.color, ff.color))
	ff.StyleOverride = &so
	windFilter := func(d game.Drawable) bool {
		if _, ok := d.(*wind); !ok {
			return false
		}
		p := d.Pos()
		s := d.Size()
		if ff.Y != p.Y {
			return false
		}

		return ff.X <= p.X && ff.X > p.X-s.X
	}
	winds := ff.Game.FilterGameObjectsByLayer(windLayer, windFilter)
	if len(winds) > 0 {
		w := winds[0]
		ff.X += w.(*wind).Speed
	}

	ff.Y += ff.speed
}

type flake struct {
	game.GameObject
	speed int
	hp    int
	color int32
	swing int
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
	for len(f.Game.FilterGameObjectsByLayer(flakeLayer, flakeFilter)) > 0 && next > f.Y {
		next--
	}

	oldY := f.Y
	blown := false

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

	winds := f.Game.FilterGameObjectsByLayer(windLayer, windFilter)
	if len(winds) > 0 {
		w := winds[0]
		f.X += w.(*wind).Speed
		blown = true
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

	if oldY != f.Y && !blown {
		// TODO f.X += f.swing
		f.swing = -1 * f.swing
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
	swing := -1
	if rand.Intn(100) > 50 {
		swing = 1
	}
	return &flake{
		GameObject: game.GameObject{
			Game:          g,
			X:             x,
			Y:             y,
			W:             1,
			H:             1,
			Sprite:        string(char),
			StyleOverride: &so,
			Layer:         flakeLayer,
		},
		color: color,
		speed: speed,
		hp:    50 + hpOffset,
		swing: swing,
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
	y := rand.Intn(g.MaxHeight - 10)
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
			Layer:     windLayer,
		},
		direction: dir,
		Speed:     speed,
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

	gg := game.NewGame(s, defStyle, w, h)

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

			x := rand.Intn(5)
			for ix := 0; ix < len(rline); ix++ {
				gap := rand.Intn(gg.MaxWidth/len(rline)) + 1
				if ix == 0 {
					gap = rand.Intn(2)
				}
				// TODO
				//if gap > 5 {
				//	gap -= 5
				//}
				x += gap
				gg.AddDrawable(newFlake(gg, x, rline[ix]))
				offset := rand.Intn(4) + 1
				if rand.Intn(10) > 5 {
					offset = -1 * offset

				}
				gg.AddDrawable(newFarFlake(gg, x+offset, rline[ix]))
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
