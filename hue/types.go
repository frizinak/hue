package hue

import (
	"errors"
	"sort"

	"github.com/amimof/huego"
)

var (
	ErrNoSuchGroup = errors.New("no such group")
	ErrNoSuchLight = errors.New("no such light")
	ErrNoSuchScene = errors.New("no such scene")
)

type Property string

const (
	ID   Property = "id"
	Name Property = "name"
)

type SceneID string

type Group struct {
	huego.Group
	lights Lights
	scenes Scenes
}

func (g *Group) Lights() Lights                          { return g.lights }
func (g *Group) Light(id int) (*Light, error)            { return g.lights.Get(id) }
func (g *Group) LightByName(name string) (*Light, error) { return g.lights.Name(name) }

func (g *Group) Scenes() Scenes                          { return g.scenes }
func (g *Group) Scene(id SceneID) (*Scene, error)        { return g.scenes.Get(id) }
func (g *Group) SceneByName(name string) (*Scene, error) { return g.scenes.Name(name) }

type Light struct {
	huego.Light
}

type Scene struct {
	huego.Scene
}

type Groups map[int]*Group

func (gs Groups) Add(groups ...*Group) {
	for _, g := range groups {
		if g == nil {
			continue
		}
		gs[g.ID] = g
	}
}

func (gs Groups) Get(id int) (*Group, error) {
	if g, ok := gs[id]; ok {
		return g, nil
	}

	return nil, ErrNoSuchGroup
}

func (gs Groups) Name(name string) (*Group, error) {
	for _, g := range gs {
		if g.Name == name {
			return g, nil
		}
	}

	return nil, ErrNoSuchGroup
}

func (gs Groups) Slice() []*Group {
	s := make([]*Group, 0, len(gs))
	for _, g := range gs {
		s = append(s, g)
	}
	return s
}

func (gs Groups) Sort(prop Property) []*Group {
	s := gs.Slice()

	var sorter func(i, j int) bool
	switch prop {
	case Name:
		sorter = func(i, j int) bool { return s[i].Name < s[j].Name }
	case ID:
		fallthrough
	default:
		sorter = func(i, j int) bool { return s[i].ID < s[j].ID }
	}

	sort.Slice(s, sorter)

	return s
}

type Scenes map[SceneID]*Scene

func (s Scenes) Get(id SceneID) (*Scene, error) {
	if sc, ok := s[id]; ok {
		return sc, nil
	}

	return nil, ErrNoSuchScene
}

func (s Scenes) Name(name string) (*Scene, error) {
	for _, sc := range s {
		if sc.Name == name {
			return sc, nil
		}
	}

	return nil, ErrNoSuchScene
}

func (s Scenes) Slice() []*Scene {
	ss := make([]*Scene, 0, len(s))
	for _, sc := range s {
		ss = append(ss, sc)
	}
	return ss
}

func (s Scenes) Sort(prop Property) []*Scene {
	sc := s.Slice()

	var sorter func(i, j int) bool
	switch prop {
	case Name:
		sorter = func(i, j int) bool { return sc[i].Name < sc[j].Name }
	case ID:
		fallthrough
	default:
		sorter = func(i, j int) bool { return sc[i].ID < sc[j].ID }
	}

	sort.Slice(sc, sorter)

	return sc
}

type Lights map[int]*Light

func (ls Lights) Add(lights ...*Light) {
	for _, l := range lights {
		if l == nil {
			continue
		}
		ls[l.ID] = l
	}
}

func (ls Lights) Get(id int) (*Light, error) {
	if l, ok := ls[id]; ok {
		return l, nil
	}

	return nil, ErrNoSuchLight
}

func (ls Lights) Name(name string) (*Light, error) {
	for _, l := range ls {
		if l.Name == name {
			return l, nil
		}
	}

	return nil, ErrNoSuchLight
}

func (ls Lights) Slice() []*Light {
	s := make([]*Light, 0, len(ls))
	for _, l := range ls {
		s = append(s, l)
	}
	return s
}

func (ls Lights) Sort(prop Property) []*Light {
	s := ls.Slice()

	var sorter func(i, j int) bool
	switch prop {
	case Name:
		sorter = func(i, j int) bool { return s[i].Name < s[j].Name }
	case ID:
		fallthrough
	default:
		sorter = func(i, j int) bool { return s[i].ID < s[j].ID }
	}

	sort.Slice(s, sorter)

	return s
}
