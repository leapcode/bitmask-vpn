// storage provides an embeddable database that bitmask uses to persist
// a series of values.
// For now, we store the following items in the database:
// - Introducer metadata.
// - Private bridges.
//
// Example Usage to initilialize the default storage struct:
//
//	InitAppStorage()
//
// To work with the initialized storage call:
//
//	GetStorage()
//

package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"

	"0xacab.org/leap/bitmask-core/pkg/models"
)

var AppName = "bitmask"
var appStorage *Storage

const (
	INTRODUCER  = "INTRODUCER"
	BRIDGE      = "BRIDGE"
	COUNTRYCODE = "COUNTRYCODE"
)

type Storage struct {
	store Store
}

// CompareIntroducer is a function type for the comparison of introducers
type CompareIntroducer func(introducer models.Introducer) bool

// CompareBridge is a function type for the comparison of bridges
type CompareBridge func(bridge models.Bridge) bool

// InitAppStorage initializes the global storage with a given storage instance
func InitAppStorageWith(store Store) {
	if appStorage != nil {
		appStorage.Close()
	}
	appStorage = NewStorageWithStore(store)
}

// Init AppStorage initializes the global storage with the default storage directory
func InitAppStorage() error {
	if appStorage != nil {
		appStorage.Close()
	}
	storage, err := NewStorageWithDefaultDir()
	if err != nil {
		return fmt.Errorf("failed to initialize app storage: %v", err)
	}
	appStorage = storage
	return nil
}

// NewStorage initializes a new storage with a given path.
// `NewStorageWithDefaultDir` should be preferred to initialize a storage, since it will try to pick a default path.
func NewStorage(path string) (*Storage, error) {
	store, err := NewDBStore(path)
	if err != nil {
		return nil, err
	}
	return &Storage{
		store: store,
	}, nil
}

// NewStorageWithStore initializes the storage struct with a custom store. This
// can be used to implement custom storage adapters e.g. for different database types
func NewStorageWithStore(store Store) *Storage {
	return &Storage{
		store: store,
	}
}

// NewStoreageWithDefaultDir initializes a storage struct with a storm DB
// and looks up the correct default paths for Linux, MacOS and Windows.
// Don't call this method in context of mobile app development
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

// AddIntroducers adds a new Introducer model from URL
// to the storage
func (s *Storage) AddIntroducer(url string) error {
	fqdn, err := getFQDNFromIntroducerURL(url)
	if err != nil {
		return err
	}

	item := &models.Introducer{
		FQDN:      *fqdn,
		URL:       url,
		CreatedAt: time.Now(),
	}

	// delete existing introducer with same FQDN as we only want
	// to allow one for per provider for now
	err = s.DeleteIntroducer(*fqdn)
	if err != nil {
		return err
	}

	introducers, err := s.getAllIntroducers()

	if err != nil {
		return err
	}
	introducers = append(introducers, *item)
	return s.saveIntroducers(introducers)
}

func (s *Storage) getAllIntroducers() ([]models.Introducer, error) {
	// Create an empty slice of Introducer
	emptySlice := []models.Introducer{}
	bytes, _ := json.Marshal(emptySlice)

	introducerString := s.store.GetByteArrayWithDefault(INTRODUCER, bytes)
	introducers, err := unmarshalJSON[[]models.Introducer](introducerString)
	if err != nil {
		return nil, err
	}
	return *introducers, err
}

func (s *Storage) saveIntroducers(introducers []models.Introducer) error {
	bytes, err := json.Marshal(introducers)
	if err != nil {
		return err
	}
	s.store.SetByteArray(INTRODUCER, bytes)
	return nil
}

func unmarshalJSON[T any](data []byte) (*T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ListIntroducers returns an array of all introducers
func (s *Storage) ListIntroducers() ([]models.Introducer, error) {
	return s.getAllIntroducers()
}

// GetIntroducerByFQDN returns the first introducer for a given fqdn.
func (s *Storage) GetIntroducerByFQDN(fqdn string) (*models.Introducer, error) {
	compare := func(intro models.Introducer) bool {
		return intro.FQDN == fqdn
	}
	return s.getFirstIntroducer(compare)
}

// GetIntroducerByURL will return the Introducer with the given URL, if found, and an error.
func (s *Storage) GetIntroducerByURL(url string) (*models.Introducer, error) {
	compare := func(intro models.Introducer) bool {
		return intro.URL == url
	}
	return s.getFirstIntroducer(compare)
}

func (s *Storage) getFirstIntroducer(compare CompareIntroducer) (*models.Introducer, error) {
	introducers, err := s.getAllIntroducers()
	if err != nil {
		return nil, err
	}
	for _, intro := range introducers {
		if compare(intro) {
			return &intro, nil // Return a pointer to the found Introducer
		}
	}
	return nil, fmt.Errorf("introducer not found")
}

func (s *Storage) updateFirstIntroducer(compare CompareIntroducer, lastUsed time.Time) ([]models.Introducer, error) {
	introducers, err := s.getAllIntroducers()
	if err != nil {
		return nil, err
	}

	found := false
	for _, intro := range introducers {
		if compare(intro) {
			intro.LastUsed = lastUsed
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("introducer not found")
	}

	return introducers, nil
}

// DeleteIntroducer deletes all introducers for a given fqdn.
func (s *Storage) DeleteIntroducer(fqdn string) error {
	if fqdn == "" {
		return errors.New("need to pass fully qualified domain name of the introducer")
	}
	var introducers, err = s.getAllIntroducers()
	if err != nil {
		return err
	}

	compare := func(intro models.Introducer) bool {
		return fqdn == intro.FQDN
	}

	updatedIntroducers := func() []models.Introducer {
		for i, intro := range introducers {
			if compare(intro) {
				// Swap with the last element
				lastIndex := len(introducers) - 1
				introducers[i] = introducers[lastIndex] // Move the last element to the current position
				return introducers[:lastIndex]          // Return the truncated slice
			}
		}
		return introducers // Return original slice if ID not found
	}()

	return s.saveIntroducers(updatedIntroducers)
}

// NewBridge creates a new Bridge from the passed parameters.
func (s *Storage) NewBridge(name, bridgeType, location, raw string) error {
	item := &models.Bridge{
		Name:     name,
		Type:     bridgeType,
		Location: location,
		Raw:      raw,
	}
	item.CreatedAt = time.Now()
	bridges, err := s.getAllBridges()
	if err != nil {
		return err
	}

	comparison := func(intro models.Bridge) bool {
		return intro.Name == name || intro.Raw == raw
	}
	bridge, _ := s.getFirstBridge(comparison)
	if bridge != nil {
		return fmt.Errorf("bridge %v already saved", bridge)
	}

	bridges = append(bridges, *item)
	return s.saveBridges(bridges)
}

func (s *Storage) getAllBridges() ([]models.Bridge, error) {
	// Create an empty slice of Introducer
	emptySlice := []models.Bridge{}
	bytes, _ := json.Marshal(emptySlice)

	bridgeBytes := s.store.GetByteArrayWithDefault(BRIDGE, bytes)
	bridges, err := unmarshalJSON[[]models.Bridge](bridgeBytes)
	if err != nil {
		return nil, err
	}
	return *bridges, err
}

func (s *Storage) saveBridges(bridges []models.Bridge) error {
	bytes, err := json.Marshal(bridges)
	if err != nil {
		return err
	}
	s.store.SetByteArray(BRIDGE, bytes)
	return nil
}

// ListBridges returns an array of all the bridges.
func (s *Storage) ListBridges() ([]models.Bridge, error) {
	return s.getAllBridges()
}

func (s *Storage) getFirstBridge(compare CompareBridge) (*models.Bridge, error) {
	bridges, err := s.getAllBridges()
	if err != nil {
		return nil, err
	}
	for _, bridge := range bridges {
		if compare(bridge) {
			return &bridge, nil // Return a pointer to the found Bridge
		}
	}
	return nil, fmt.Errorf("bridge not found")
}

func (s *Storage) getBridges(compare CompareBridge) ([]models.Bridge, error) {
	bridges, err := s.getAllBridges()
	if err != nil {
		return nil, err
	}
	var result []models.Bridge

	for _, bridge := range bridges {
		if compare(bridge) {
			result = append(result, bridge)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("bridge not found")
	}

	return result, nil
}

// GetBridgeByName will return the Bridge with the given Name, if found, and an error.
func (s *Storage) GetBridgeByName(name string) (*models.Bridge, error) {
	compare := func(bridge models.Bridge) bool {
		return bridge.Name == name
	}
	return s.getFirstBridge(compare)
}

// GetBridgesByType will return all Bridges with the given Type, if found, and an error.
func (s *Storage) GetBridgesByType(bridgeType string) ([]models.Bridge, error) {
	compare := func(bridge models.Bridge) bool {
		return bridge.Type == bridgeType
	}
	return s.getBridges(compare)
}

// GetBridgesByLocation will return all Bridges with the given Location, if found, and an error.
func (s *Storage) GetBridgesByLocation(location string) ([]models.Bridge, error) {
	compare := func(bridge models.Bridge) bool {
		return bridge.Location == location
	}
	return s.getBridges(compare)
}

// DeleteBridge accepts a name and an ID. If you want to delete by name, pass 0 as the ID;
// if you want to delete by ID, pass the empty string as name.
func (s *Storage) DeleteBridge(name string) error {
	bridges, err := s.getAllBridges()
	if err != nil {
		return err
	}

	compare := func(bridge models.Bridge) bool {
		return bridge.Name == name
	}

	updatedBridges := func() []models.Bridge {
		for i, bridge := range bridges {
			if compare(bridge) {
				// Swap with the last element
				lastIndex := len(bridges) - 1
				bridges[i] = bridges[lastIndex] // Move the last element to the current position
				return bridges[:lastIndex]      // Return the truncated slice
			}
		}
		return bridges // Return original slice if ID not found
	}()

	return s.saveBridges(updatedBridges)

}

// Close closes the db connection
func (s *Storage) Close() {
	_ = s.store.Close()
}

// GetStorage returns an initialized storage instance
// if the global appStorage was initialized it has precedence
func GetStorage() (*Storage, error) {
	var storage *Storage
	var err error
	if appStorage == nil {
		storage, err = NewStorageWithDefaultDir()
		if err != nil {
			return nil, err
		}
	} else {
		storage = appStorage
	}
	return storage, nil

}

func getFQDNFromIntroducerURL(introducerURL string) (*string, error) {
	// Parse the introducer URL string
	u, err := url.Parse(introducerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse introducer URL: %w", err)
	}

	// Extract FQDN from query parameters
	fqdn := u.Query().Get("fqdn")
	if fqdn == "" {
		return nil, fmt.Errorf("FQDN not found in the introducer URL")
	}

	return &fqdn, nil
}

// UpdateLastUsedForIntroducer will attempt to update the LastUsed timestamp for the introducer
// that matches the passed URL.
func (s *Storage) UpdateLastUsedForIntroducer(url string) error {
	var compare CompareIntroducer = func(introducer models.Introducer) bool {
		return introducer.URL == url
	}
	introducers, err := s.updateFirstIntroducer(compare, time.Now())
	if err != nil {
		return err
	}

	return s.saveIntroducers(introducers)
}

func (s *Storage) SaveFallbackCountryCode(cc string) {
	s.store.SetString(COUNTRYCODE, cc)
}

func (s *Storage) GetFallbackCountryCode() string {
	return s.store.GetString(COUNTRYCODE)
}
