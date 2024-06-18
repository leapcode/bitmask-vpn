// storage provides an embeddable database that bitmask uses to persist
// a series of values.
// For now, we store the following items in the database:
// - Introducer metadata.
// - Private bridges.
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/asdine/storm/v3"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog/log"

	"0xacab.org/leap/bitmask-core/pkg/models"
)

var AppName = "bitmask"

type Storage struct {
	db *storm.DB
}

// NewStorage initializes a new storage with a given path. `NewStorageWithDefaultDir` should be preferred to initialize a storage, since it will try to pick a default path.
func NewStorage(path string) (*Storage, error) {
	fp := filepath.Join(path, "bitmask.db")
	db, err := storm.Open(fp)
	if err != nil {
		return nil, err
	}

	// Initialize buckets and indexes
	if err := db.Init(&models.Introducer{}); err != nil {
		return nil, err
	}
	if err := db.Init(&models.Bridge{}); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}

func NewStorageWithDefaultDir() (*Storage, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	var configPath string
	switch _os := os.Getenv("GOOS"); _os {
	case "windows":
		configPath = filepath.Join(os.Getenv("APPDATA"), AppName)
	default:
		// This will cover both 'darwin' (macOS) and 'linux'
		configPath = filepath.Join(home, ".config", AppName)
	}

	err = os.MkdirAll(configPath, 0700)
	if err != nil {
		return nil, err
	}

	return NewStorage(configPath)
}

// NewIntroducer creates a new Introducer from the name and URL.
func (s *Storage) NewIntroducer(name, url string) error {
	item := &models.Introducer{
		Name: name,
		URL:  url,
	}
	item.CreatedAt = time.Now()
	return s.db.Save(item)
}

// ListIntroducers returns an array of all the introducers
func (s *Storage) ListIntroducers() ([]models.Introducer, error) {
	var items []models.Introducer
	err := s.db.AllByIndex("CreatedAt", &items)
	return items, err
}

// GetIntroducerByID will return the Introducer with the given ID, if found, and an error.
func (s *Storage) GetIntroducerByID(id int) (models.Introducer, error) {
	var introducer models.Introducer
	err := s.db.One("ID", id, &introducer)
	return introducer, err
}

// GetIntroducerByName will return the Introducer with the given Name, if found, and an error.
func (s *Storage) GetIntroducerByName(name string) (models.Introducer, error) {
	var introducer models.Introducer
	err := s.db.One("Name", name, &introducer)
	return introducer, err
}

// DeleteIntroducer accepts a name and an ID. If you want to delete by name, pass 0 as the ID;
// if you want to delete by ID, pass the empty string as name.
func (s *Storage) DeleteIntroducer(id int, name string) error {
	if id == 0 && name == "" {
		return fmt.Errorf("need to pass id or name")
	}
	var introducer models.Introducer
	var err error
	switch {
	case id != 0:
		introducer, err = s.GetIntroducerByID(id)
		if err != nil {
			return err
		}
	case name != "":
		introducer, err = s.GetIntroducerByName(name)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled case")
	}
	return s.db.DeleteStruct(&introducer)
}

// TODO GetIntroducersNeverUsed - useful for prune

// NewBridge creates a new Bridge from the passed parameters.
func (s *Storage) NewBridge(name, bridgeType, location, raw string) error {
	item := &models.Bridge{
		Name:     name,
		Type:     bridgeType,
		Location: location,
		Raw:      raw,
	}
	item.CreatedAt = time.Now()
	return s.db.Save(item)
}

// ListBridges returns an array of all the bridges.
func (s *Storage) ListBridges() ([]models.Bridge, error) {
	var items []models.Bridge
	err := s.db.AllByIndex("CreatedAt", &items)
	return items, err
}

// GetBridgeByID will return the Bridge with the given ID, if found, and an error.
func (s *Storage) GetBridgeByID(id int) (models.Bridge, error) {
	var bridge models.Bridge
	err := s.db.One("ID", id, &bridge)
	return bridge, err
}

// GetBridgeByName will return the Bridge with the given Name, if found, and an error.
func (s *Storage) GetBridgeByName(name string) (models.Bridge, error) {
	var bridge models.Bridge
	err := s.db.One("Name", name, &bridge)
	return bridge, err
}

// GetBridgesByType will return all Bridges with the given Type, if found, and an error.
func (s *Storage) GetBridgesByType(bridgeType string) ([]models.Bridge, error) {
	var bridges []models.Bridge
	err := s.db.Find("Type", bridgeType, &bridges)
	return bridges, err
}

// GetBridgesByLocation will return all Bridges with the given Location, if found, and an error.
func (s *Storage) GetBridgesByLocation(location string) ([]models.Bridge, error) {
	var bridges []models.Bridge
	err := s.db.Find("Location", location, &bridges)
	return bridges, err
}

// DeleteBridge accepts a name and an ID. If you want to delete by name, pass 0 as the ID;
// if you want to delete by ID, pass the empty string as name.
func (s *Storage) DeleteBridge(id int, name string) error {
	if id == 0 && name == "" {
		return fmt.Errorf("need to pass id or name")
	}
	var bridge models.Bridge
	var err error
	switch {
	case id != 0:
		bridge, err = s.GetBridgeByID(id)
		if err != nil {
			return err
		}
	case name != "":
		bridge, err = s.GetBridgeByName(name)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unhandled case")
	}
	return s.db.DeleteStruct(&bridge)
}

// Close closes the db connection
func (s *Storage) Close() {
	s.db.Close()
}

// MaybeUpdateLastUsedForIntroducer will attempt to update the LastUsed timestamp for the introducer
// that matches the passed URL.
func MaybeUpdateLastUsedForIntroducer(url string) error {
	db, err := NewStorageWithDefaultDir()
	if err != nil {
		return err
	}
	defer db.Close()

	var intro models.Introducer
	err = db.db.One("URL", url, &intro)
	if err != nil {
		log.Error().Err(err)
		return err
	}
	intro.LastUsed = time.Now()
	err = db.db.Save(&intro)
	if err != nil {
		return err
	}
	return nil
}
