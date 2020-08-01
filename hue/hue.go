package hue

import (
	"fmt"
	"image/color"
	"strconv"

	"github.com/amimof/huego"
)

type App struct {
	bridge *huego.Bridge
	groups Groups
	lights Lights
}

func New(ip, pass string) *App {
	bridge := huego.New(ip, pass)
	return &App{bridge: bridge}
}

func (a *App) Init() error {
	errs := make(chan error, 2)

	var gs []huego.Group
	var lights Lights

	go func() {
		var err error
		gs, err = a.bridge.GetGroups()
		errs <- err
	}()

	go func() {
		ls, err := a.bridge.GetLights()
		errs <- err

		if err == nil {
			lights = make(Lights, len(ls))
			for _, l := range ls {
				lights[l.ID] = &Light{Light: l}
			}
		}
	}()

	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			return err
		}
	}

	groups := make(Groups, len(gs))
	for _, g := range gs {
		n := make(Lights, len(g.Lights))
		groups[g.ID] = &Group{Group: g, lights: n, scenes: make(Scenes)}

		for _, _lid := range g.Lights {
			lid, err := strconv.Atoi(_lid)
			if err != nil {
				return err
			}
			if _, ok := lights[lid]; !ok {
				return fmt.Errorf("missing light %d", lid)
			}
			n[lid] = lights[lid]
		}
	}

	a.groups = groups
	a.lights = lights

	scenes, err := a.bridge.GetScenes()
	if err != nil {
		return err
	}

	for _, scene := range scenes {
		if scene.Group == "" {
			continue
		}
		groupID, err := strconv.Atoi(scene.Group)
		if err != nil {
			return fmt.Errorf("failed to parse group id %s for scene %s: %w", scene.Group, scene.Name, err)
		}
		if _, ok := a.groups[groupID]; !ok {
			continue
		}
		a.groups[groupID].scenes[SceneID(scene.ID)] = &Scene{scene}
	}

	return nil
}

func (a *App) Groups() Groups                          { return a.groups }
func (a *App) Group(id int) (*Group, error)            { return a.groups.Get(id) }
func (a *App) GroupByName(name string) (*Group, error) { return a.groups.Name(name) }

func (a *App) Lights() Lights                          { return a.lights }
func (a *App) Light(id int) (*Light, error)            { return a.lights.Get(id) }
func (a *App) LightByName(name string) (*Light, error) { return a.lights.Name(name) }

func (a *App) SetColor(ls []*Light, c color.Color) error {
	hsv := ColorHSV(c)
	if hsv.Hue == 0 {
		hsv.Hue = 1<<16 - 1
	}
	if hsv.Sat == 0 {
		hsv.Sat = 1
	}
	if hsv.Value == 0 {
		for _, l := range ls {
			if err := l.Off(); err != nil {
				return err
			}
		}
		return nil
	}

	for _, l := range ls {
		a.bridge.SetLightState(
			l.ID,
			huego.State{
				On:  true,
				Hue: hsv.Hue,
				Sat: hsv.Sat,
			},
		)

		switch l.State.ColorMode {
		case "ct":
			ct := ColorTemp(c)
			if ct.T == 0 {
				ct.T = 1
			}
			state := huego.State{On: true, Ct: ct.T}
			if _, err := a.bridge.SetLightState(l.ID, state); err != nil {
				return err
			}

		case "xy":
			xy := ColorXY(c, ProfileWideGamut)
			state := huego.State{On: true, Xy: []float32{xy.X, xy.Y}}
			if _, err := a.bridge.SetLightState(l.ID, state); err != nil {
				return err
			}
		}

		if err := l.Bri(hsv.Value); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) SetScene(id SceneID) error {
	for _, g := range a.groups {
		for sid := range g.scenes {
			if sid == id {
				return g.Group.Scene(string(id))
			}
		}
	}

	return ErrNoSuchScene
}

func (a *App) SetGroupSceneByName(groups Groups, name string) error {
	for _, g := range groups {
		for sid, s := range g.scenes {
			if s.Name == name {
				return g.Group.Scene(string(sid))
			}
		}
	}

	return ErrNoSuchScene
}
