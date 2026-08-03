package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jwalton/go-supportscolor"
	"github.com/xypwn/filediver/app"
	"github.com/xypwn/filediver/app/appconfig"
	"github.com/xypwn/filediver/config"
	"github.com/xypwn/filediver/hashes"
	"github.com/xypwn/filediver/stingray"
	"github.com/xypwn/filediver/stingray/level"
	"github.com/xypwn/filediver/stingray/prefab"
)

func dumpLevelObjectNames(a *app.App, fileID stingray.FileID) error {
	bs, err := a.DataDir.Read(fileID, stingray.DataMain)
	if err != nil {
		return err
	}
	levelData, err := level.LoadLevel(bytes.NewReader(bs), nil)
	if err != nil {
		return err
	}
	for _, unit := range levelData.Units {
		knownName, ok := a.Hashes[unit.Name]
		if ok {
			fmt.Println(knownName)
		} else {
			fmt.Println(unit.Name.String())
		}
	}
	for _, extra := range levelData.UnkExtraUnitContainers {
		for _, unit := range extra.ExtraUnits {
			knownName, ok := a.Hashes[unit.Name]
			if ok {
				fmt.Println(knownName)
			} else {
				fmt.Println(unit.Name.String())
			}
		}
	}
	return nil
}

func dumpPrefabObjectNames(a *app.App, fileID stingray.FileID) error {
	bs, err := a.DataDir.Read(fileID, stingray.DataMain)
	if err != nil {
		return err
	}
	prefabData, err := prefab.Load(bytes.NewReader(bs))
	if err != nil {
		return err
	}
	for _, unit := range prefabData.Units {
		knownName, ok := a.Hashes[unit.Name]
		if ok {
			fmt.Println(knownName)
		} else {
			fmt.Println(unit.Name.String())
		}
	}
	for _, prefab := range prefabData.NestedPrefabs {
		knownName, ok := a.Hashes[prefab.Name]
		if ok {
			fmt.Println(knownName)
		} else {
			fmt.Println(prefab.Name.String())
		}
	}
	return nil
}

func main() {
	prt := app.NewConsolePrinter(
		supportscolor.Stderr().SupportsColor,
		os.Stderr,
		os.Stderr,
	)

	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	gameDir, err := app.DetectGameDir()
	if err != nil {
		prt.Fatalf("Unable to detect game install directory.")
	}

	knownHashes := app.ParseHashes(hashes.Hashes)
	knownThinHashes := app.ParseHashes(hashes.ThinHashes)

	a, err := app.OpenGameDir(ctx, gameDir, knownHashes, knownThinHashes, stingray.ThinHash{}, func(curr int, total int) {
		prt.Statusf("Opening game directory %.0f%%", float64(curr)/float64(total)*100)
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			prt.NoStatus()
			prt.Warnf("Level name dump canceled")
			return
		} else {
			prt.Fatalf("%v", err)
		}
	}
	prt.NoStatus()

	files, err := a.MatchingFiles("", "", []string{"prefab", "level"}, nil, "", prt.Infof)
	if err != nil {
		prt.Fatalf("%v", err)
	}

	var cfg appconfig.Config
	config.InitDefault(&cfg)
	count := 0
	for fileID := range files {
		prt.Statusf("File %v/%v - %v.%v", count+1, len(files), fileID.Name.String(), a.LookupHash(fileID.Type))
		dumpNames := dumpLevelObjectNames
		if fileID.Type == stingray.Sum("prefab") {
			dumpNames = dumpPrefabObjectNames
		} else if fileID.Type != stingray.Sum("level") {
			continue
		}
		if err := dumpNames(a, fileID); err != nil {
			if errors.Is(err, context.Canceled) {
				prt.NoStatus()
				prt.Warnf("Name dump canceled, exiting cleanly")
				return
			} else {
				prt.Errorf("%v", err)
			}
		}
		count++
	}
}
