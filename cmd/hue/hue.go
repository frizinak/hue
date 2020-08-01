package main

import (
	"errors"
	"fmt"
	"image/color"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/amimof/huego"
	"github.com/frizinak/hue/app"
	"github.com/frizinak/hue/hue"
)

func temp(app *hue.App, ls []*hue.Light, min, max int) error {
	paths, err := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	if err != nil {
		return err
	}

	last := color.RGBA{}
	c := color.RGBA{A: 255}
	for {
		var cur int
		for _, p := range paths {
			d, err := ioutil.ReadFile(p)
			if err != nil {
				return err
			}

			n, err := strconv.Atoi(strings.TrimSpace(string(d)))
			if err != nil {
				return err
			}

			if n > cur {
				cur = n
			}
		}

		time.Sleep(time.Millisecond * 80)
		scale := float64(cur/1000-min) / float64(max-min)
		if scale > 1.0 {
			scale = 1.0
		}

		var r, g float64
		switch {
		case scale < 0.2:
			r, g = 0.0, 1.0
		case scale < 0.5:
			r, g = 0.5, 0.5
		default:
			r = scale
			g = 1.0 - r
		}

		c.R = uint8(r * float64(255))
		c.G = uint8(g * float64(255))
		if c == last {
			continue
		}

		last = c

		if err := app.SetColor(ls, c); err != nil {
			return err
		}
	}
}

func exit(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func ensureFile(file string, defaultContents []byte) error {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		if defaultContents != nil {
			_, err = f.Write(defaultContents)
		}
		return err
	}

	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func main() {
	const ns = "hue.frizinak"

	configDir, err := os.UserConfigDir()
	exit(err)
	configDir = filepath.Join(configDir, ns)
	os.MkdirAll(configDir, 0700)

	ipFile := filepath.Join(configDir, "ip")
	passFile := filepath.Join(configDir, "pass")
	exit(ensureFile(ipFile, nil))
	exit(ensureFile(passFile, nil))
	ipData, err := ioutil.ReadFile(ipFile)
	exit(err)
	passData, err := ioutil.ReadFile(passFile)
	exit(err)

	ok := true
	haveIP := true
	if len(passData) == 0 {
		ok = false
		fmt.Fprintf(os.Stderr, "'%s' is empty\n", passFile)
	}
	if len(ipData) < 7 {
		haveIP = false
		ok = false
		fmt.Fprintf(os.Stderr, "'%s' is empty\n", ipFile)
	}

	if !ok {
		var ip = string(ipData)
		var hub *huego.Bridge
		if !haveIP {
			fmt.Println("discovering hue bridge, you can cancel this operation and fill in the above files if you already have its ip and an app password")
			var err error
			hub, err = huego.Discover()
			if err != nil {
				exit(fmt.Errorf("Something went wrong during discover: %w", err))
			}
			ip = hub.Host
			fmt.Printf("found hub at '%s'\n", ip)
		}

		fmt.Printf("creating new app on bridge at ip %s\n", ip)
		if hub == nil {
			hub = huego.New(ip, "")
		}
		fmt.Println("go press the bridge button and press enter to continue")
		fmt.Scanln()
		pass, err := hub.CreateUser(ns)
		exit(err)
		exit(ioutil.WriteFile(ipFile, []byte(hub.Host), 0600))
		exit(ioutil.WriteFile(passFile, []byte(pass), 0600))
		fmt.Println("success")
		os.Exit(0)
	}

	happ := hue.New(string(ipData), string(passData))

	exit(happ.Init())
	app := app.New(happ)

	if len(os.Args) < 2 {
		exit(errors.New("no command given"))
	}

	switch os.Args[1] {
	// case "temp":
	// 	if len(os.Args) < 3 {
	// 		exit(errors.New("missing argument lights/groups"))
	// 	}
	// 	lights, err := lightIDs(app, os.Args[2])
	// 	exit(err)
	// 	exit(temp(app, lights.Slice(), 50, 80))

	case "help", "-h", "--help":
		fmt.Println("hue")
		fmt.Println(" list, ls:                               list lights")
		fmt.Println(" list-groups, lsg:                       list groups")
		fmt.Println(" set <scene-id>:                         set a scene")
		fmt.Println(" set <entities> <color>:                 set comma separated list of ENTITY to a color")
		fmt.Println(" set <group-id|group-name> <scene-name>: set a scene")
		fmt.Println()
		fmt.Println("ENTITY:")
		fmt.Println(" group-name / group-id  e.g.: 'LivingRoom' / 'g1'")
		fmt.Println(" light-name / light-id  e.g.: 'Spot 1' / 'l2'")
		os.Exit(0)

	case "list", "ls":
		var filter string
		if len(os.Args) > 2 {
			filter = os.Args[2]
		}

		for _, l := range happ.Lights().Sort(hue.ID) {
			if filter != "" {
				name := strings.ToLower(l.Name)
				filter = strings.ToLower(filter)
				if !strings.Contains(name, filter) {
					continue
				}
			}
			fmt.Printf("%02d] %s\n", l.ID, l.Name)
		}
	case "list-groups", "lsg":
		var filter string
		if len(os.Args) > 2 {
			filter = os.Args[2]
		}

		for _, g := range happ.Groups().Sort(hue.ID) {
			if filter != "" {
				name := strings.ToLower(g.Name)
				filter = strings.ToLower(filter)
				if !strings.Contains(name, filter) {
					continue
				}
			}
			fmt.Printf("%-3s) %-30s\n", app.GID(g), g.Name)
			for _, l := range g.Lights().Sort(hue.ID) {
				fmt.Printf("    %-3s) %-30s\n", app.LID(l), l.Name)
			}

			for _, s := range g.Scenes().Sort(hue.Name) {
				fmt.Printf("    %-15s) %-30s\n", app.ScID(s), s.Name)
			}
		}

	case "set":
		if len(os.Args) == 3 {
			exit(happ.SetScene(hue.SceneID(os.Args[2])))
			return
		}

		if len(os.Args) < 4 {
			exit(errors.New("missing argument lights/groups and/or color"))
		}
		c, err := app.HexColor(os.Args[3])
		if err == nil {
			lights, err := app.Lights(os.Args[2])
			exit(err)
			exit(happ.SetColor(lights.Slice(), c))
			return
		}

		groups, lights, err := app.ParseIdentifiers(os.Args[2])
		exit(err)
		if len(lights) != 0 {
			exit(errors.New("specified lights, only supports groups"))
		}

		exit(happ.SetGroupSceneByName(groups, os.Args[3]))

	default:
		exit(errors.New("no such command"))
	}
}
