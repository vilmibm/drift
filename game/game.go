package game

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type Game struct {
	debug        bool
	Screen       tcell.Screen
	DefaultStyle tcell.Style
	Style        tcell.Style
	MaxWidth     int
	MaxHeight    int
	layers       map[int][]Drawable
	// Logger       *log.Logger
}

func NewGame(screen tcell.Screen, style tcell.Style, maxWidth, maxHeight int) *Game {
	return &Game{
		Screen:    screen,
		Style:     style,
		MaxWidth:  maxWidth,
		MaxHeight: maxHeight,
		layers:    map[int][]Drawable{},
	}
}

func (g *Game) DrawStr(x, y int, str string, style *tcell.Style) {
	var st tcell.Style
	if style == nil {
		st = g.DefaultStyle
	} else {
		st = *style
	}
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		g.Screen.SetContent(x, y, c, comb, st)
		x += w
	}
}

func (g *Game) AddDrawable(d Drawable) {
	l := d.GetLayer()
	if _, ok := g.layers[l]; !ok {
		g.layers[l] = []Drawable{}
	}
	g.layers[l] = append(g.layers[l], d)
}

func (g *Game) Destroy(d Drawable) {
	ds := g.layers[d.GetLayer()]
	newDrawables := []Drawable{}
	for _, dd := range ds {
		if dd == d {
			continue
		}
		newDrawables = append(newDrawables, dd)
	}
	g.layers[d.GetLayer()] = newDrawables
}

func (g *Game) Update() {
	for _, ds := range g.layers {
		for _, gobj := range ds {
			gobj.Update()
		}
	}
}

func (g *Game) Draw() {
	lix := []int{}
	for k := range g.layers {
		lix = append(lix, k)
	}
	sort.Ints(lix)
	for _, l := range lix {
		for _, d := range g.layers[l] {
			d.Draw()
		}

	}
}

func (g *Game) FindGameObject(fn func(Drawable) bool) Drawable {
	for _, ds := range g.layers {
		for _, d := range ds {
			if fn(d) {
				return d
			}
		}
	}

	return nil
}

func (g *Game) FilterGameObjects(fn func(Drawable) bool) []Drawable {
	out := []Drawable{}
	for _, ds := range g.layers {
		for _, d := range ds {
			if fn(d) {
				out = append(out, d)
			}
		}
	}
	return out
}

func (g *Game) FilterGameObjectsByLayer(layer int, fn func(Drawable) bool) []Drawable {
	out := []Drawable{}
	ds, ok := g.layers[layer]
	if !ok {
		return out
	}

	for _, d := range ds {
		if fn(d) {
			out = append(out, d)
		}
	}

	return out
}

type Drawable interface {
	Draw()
	Update()
	Pos() Point
	Size() Point
	Transform(int, int)
	GetLayer() int
}

type GameObject struct {
	X             int
	Y             int
	W             int
	H             int
	Sprite        string
	Game          *Game
	StyleOverride *tcell.Style
	// TODO and thus, Drawable becomes a misnomer
	Invisible bool
	Layer     int
}

func (g *GameObject) Update() {}

func (g *GameObject) Transform(x, y int) {
	g.X += x
	g.Y += y
}

func (g *GameObject) Pos() Point {
	return Point{g.X, g.Y}
}

func (g *GameObject) Size() Point {
	return Point{g.W, g.H}
}

func (g *GameObject) GetLayer() int {
	return g.Layer
}

func (g *GameObject) Draw() {
	if g.Invisible {
		return
	}
	var style *tcell.Style
	if g.StyleOverride != nil {
		style = g.StyleOverride
	}
	lines := strings.Split(g.Sprite, "\n")
	for i, line := range lines {
		l := line
		w := runewidth.StringWidth(line)
		if g.X+w > g.Game.MaxWidth {
			space := g.Game.MaxWidth - g.X
			comb := []rune{}
			for i, r := range line {
				if i > space {
					break
				}
				comb = append(comb, r)
			}
			l = string(comb)
		}
		g.Game.DrawStr(g.X, g.Y+i, l, style)
	}
}

type Point struct {
	X int
	Y int
}

func (p Point) String() string {
	return fmt.Sprintf("<%d, %d>", p.X, p.Y)
}

func (p Point) Equals(o Point) bool {
	return p.X == o.X && p.Y == o.Y
}

type Ray struct {
	Points []Point
}

func NewRay(a Point, b Point) *Ray {
	r := &Ray{
		Points: []Point{},
	}

	if a.Equals(b) {
		return r
	}

	xDir := 1
	if a.X > b.X {
		xDir = -1
	}
	yDir := 1
	if a.Y > b.Y {
		yDir = -1
	}

	x := a.X
	y := a.Y

	for x != b.X || y != b.Y {
		r.AddPoint(x, y)
		if x != b.X {
			x += xDir * 1
		}
		if y != b.Y {
			y += yDir * 1
		}
	}

	r.AddPoint(x, y)

	return r
}

func (r *Ray) AddPoint(x, y int) {
	r.Points = append(r.Points, Point{X: x, Y: y})
}

func (r *Ray) Length() int {
	return len(r.Points)
}
