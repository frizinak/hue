package app

import (
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/frizinak/hue/hue"
)

type App struct {
	app *hue.App
}

func New(hue *hue.App) *App {
	return &App{hue}
}

func (a *App) GID(g *hue.Group) string  { return fmt.Sprintf("g%d", g.ID) }
func (a *App) LID(l *hue.Light) string  { return fmt.Sprintf("l%d", l.ID) }
func (a *App) ScID(s *hue.Scene) string { return s.ID }

func (a *App) EntityID(id string) (*hue.Group, *hue.Light, error) {
	t := id[0]
	n, err := strconv.Atoi(id[1:])
	if err != nil {
		return nil, nil, errors.New("invalid identifier: not an integer")
	}

	switch t {
	case 'g':
		g, err := a.app.Group(n)
		return g, nil, err
	case 'l':
		l, err := a.app.Light(n)
		return nil, l, err
	default:
		return nil, nil, errors.New("invalid identifier: no such type")
	}
}

func (a *App) EntityNames(names []string) (hue.Groups, hue.Lights, []string, error) {
	lights := a.app.Lights().Slice()
	groups := a.app.Groups().Slice()

	ml := make(hue.Lights)
	mg := make(hue.Groups)

	unused := make([]string, 0)

	for _, p := range names {
		found := false
		for i, l := range lights {
			if l.Name == p {
				found = true
				ml[i] = l
			}
		}
		for i, g := range groups {
			if g.Name == p {
				found = true
				mg[i] = g
			}
		}
		if !found {
			unused = append(unused, p)
		}
	}

	return mg, ml, unused, nil
}

func (a *App) EntityIDs(ids []string) (hue.Groups, hue.Lights, []string, error) {
	list := make([]string, 0, len(ids))
	unused := make([]string, 0)
	var lastType byte
	for _, p := range ids {
		p = strings.TrimSpace(p)
		rn := strings.SplitN(p, "-", 2)
		if len(rn) == 1 {
			_, err := strconv.Atoi(p)
			if err != nil {
				list = append(list, p)
				lastType = p[0]
				continue
			}

			list = append(list, fmt.Sprintf("%s%s", string(lastType), p))
			continue
		}

		for i, p := range rn {
			rn[i] = strings.TrimSpace(p)
			if rn[i] == "" {
				continue
			}
		}

		from, err := strconv.Atoi(rn[0])
		if err != nil {
			from, err = strconv.Atoi(rn[0][1:])
			if err != nil {
				unused = append(unused, p)
				continue
			}
			lastType = rn[0][0]
		}

		till, err := strconv.Atoi(rn[1])
		if err != nil {
			till, err = strconv.Atoi(rn[1][1:])
			if err != nil {
				unused = append(unused, p)
			}
		}

		for i := from; i < till+1; i++ {
			list = append(list, fmt.Sprintf("%s%d", string(lastType), i))
		}
	}

	lights := make(hue.Lights)
	groups := make(hue.Groups)
	for _, i := range list {
		g, l, err := a.EntityID(i)
		if err != nil {
			unused = append(unused, i)
			continue
		}
		lights.Add(l)
		groups.Add(g)
	}

	return groups, lights, unused, nil
}

func (a *App) ParseIdentifiers(names string) (hue.Groups, hue.Lights, error) {
	unused := list(names)
	g1, l1, u1, err := a.EntityIDs(unused)
	if err != nil {
		return nil, nil, err
	}
	g2, l2, u2, err := a.EntityNames(u1)
	if err != nil {
		return nil, nil, err
	}
	if len(u2) != 0 {
		return nil, nil, fmt.Errorf("some entries could not be parsed: '%s'", strings.Join(u2, ", "))
	}
	for _, g := range g2 {
		g1.Add(g)
	}
	for _, l := range l2 {
		l1.Add(l)
	}

	return g1, l1, nil
}

func (a *App) Lights(names string) (hue.Lights, error) {
	gs, ls, err := a.ParseIdentifiers(names)
	if err != nil {
		return nil, err
	}
	for _, g := range gs {
		for _, l := range g.Lights() {
			ls.Add(l)
		}
	}

	return ls, nil
}

func (a *App) HexColor(str string) (color.Color, error) {
	c := color.RGBA{A: 255}
	if str[0] == '#' {
		str = str[1:]
	}

	var err error
	switch len(str) {
	case 8:
		_, err = fmt.Sscanf(str, "%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	case 6:
		_, err = fmt.Sscanf(str, "%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(str, "%1x%1x%1x%1x", &c.R, &c.G, &c.B, &c.A)
		c.R, c.G, c.B, c.A = c.R*17, c.G*17, c.B*17, c.A*17
	case 3:
		_, err = fmt.Sscanf(str, "%1x%1x%1x", &c.R, &c.G, &c.B)
		c.R, c.G, c.B = c.R*17, c.G*17, c.B*17
	default:
		err = errors.New("invalid hex color")
	}

	return c, err
}

func list(list string) []string {
	ps := strings.Split(list, ",")
	l := make([]string, 0, len(ps))
	for _, p := range ps {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		l = append(l, p)
	}
	return l
}
