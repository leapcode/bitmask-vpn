package backend

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"os"
	"path/filepath"
	"strconv"

	"0xacab.org/leap/bitmask-core/pkg/bootstrap"
	"0xacab.org/leap/bitmask-core/pkg/introducer"
	"0xacab.org/leap/bitmask-core/pkg/storage"
	"0xacab.org/leap/bitmask-vpn/pkg/bitmask"
	"0xacab.org/leap/bitmask-vpn/pkg/config"
	"github.com/rs/zerolog/log"
)

func fetchProviderOptsWitBootstrapper(providerURL string) *bitmask.ProviderOpts {
	cfg, err := bootstrap.NewConfigFromURL(providerURL)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to initialize bitmask-core config")
		return &bitmask.ProviderOpts{}
	}

	bootstrapper, err := bootstrap.NewAPI(cfg)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to initialize bitmask-core bootstrapper")
		return &bitmask.ProviderOpts{}
	}

	providerInfo, err := bootstrapper.GetProvider()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to fetch provider info")
		return &bitmask.ProviderOpts{}
	}

	apiVersion, _ := strconv.Atoi(providerInfo.APIVersion)
	name, ok := providerInfo.Name["en"]
	if !ok {
		name = fmt.Sprintf("provider_generic")
	}

	var caCert = []byte{}
	if providerInfo.CaCertURI != "" {
		resp, err := http.Get(providerInfo.CaCertURI)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("unable to fetch cacert")
		}
		caCert, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("unable to fetch cacert")
		}
	}
	// convert provider model to bitmask provider struct
	providerOpts := bitmask.ProviderOpts{
		AppName:         "Bitmask",
		BinaryName:      "bitmask-vpn",
		ApiURL:          providerInfo.APIURI,
		ProviderURL:     providerInfo.Domain,
		TosURL:          providerInfo.TosURL,
		ApiVersion:      apiVersion,
		AskForDonations: providerInfo.AskForDonations,
		Auth:            "annon",
		DonateURL:       providerInfo.DonateURL,
		HelpURL:         providerInfo.InfoURL,
		Provider:        name,
		CaCert:          string(caCert),
		GeolocationURL:  "",
	}

	return &providerOpts
}

func writeProviderJSONToFile(opts *bitmask.ProviderOpts, path string) error {
	data, err := json.MarshalIndent(opts, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func getProviderJSONPath(provider string) string {
	return filepath.Join(config.Path, fmt.Sprintf("provider_%s.json", provider))
}

func fetchProviderOptsWithIntroducer(introducerURL string) *bitmask.ProviderOpts {
	err := storage.InitAppStorage()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to initialize bitmask-core storage")
		return &bitmask.ProviderOpts{}
	}
	store, err := storage.GetStorage()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Got error trying to access the introducer storage")
	}

	// parse introducer URL and get the FQDN to init the bootstrap config
	intr, err := introducer.NewIntroducerFromURL(introducerURL)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Got error trying to parse the introducer URL")
	}
	_, err = store.GetIntroducerByFQDN(intr.FQDN)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Introducer doesn't exist in storage adding it")
		err = store.AddIntroducer(intr)
		if err != nil {
			log.Warn().
				Err(err).
				Str("introducerURL", introducerURL).
				Msg("Failed to add introducer to storage")
		} else {
			log.Info().Msg("Added introducer to database")
		}
	} else {
		log.Info().Msg("Found introducer in database")
	}

	cfg, err := bootstrap.NewConfigFromURL(intr.FQDN)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to create bootstrap to config")
	}

	cfg.Introducer = introducerURL
	bootstrapper, err := bootstrap.NewAPI(cfg)
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to initialize bitmask-core bootstrapper")
		return &bitmask.ProviderOpts{}
	}

	providerInfo, err := bootstrapper.GetProvider()
	if err != nil {
		log.Warn().
			Err(err).
			Msg("unable to fetch provider info")
		return &bitmask.ProviderOpts{}
	}

	apiVersion, _ := strconv.Atoi(providerInfo.APIVersion)
	name, ok := providerInfo.Name["en"]
	if !ok {
		name = fmt.Sprintf("provider_generic")
	}

	var caCert = []byte{}
	if providerInfo.CaCertURI != "" {
		resp, err := http.Get(providerInfo.CaCertURI)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("unable to fetch cacert")
		}
		caCert, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Warn().
				Err(err).
				Msg("unable to fetch cacert")
		}
	}
	// convert provider model to bitmask provider struct
	providerOpts := bitmask.ProviderOpts{
		AppName:         "Bitmask",
		BinaryName:      "bitmask-vpn",
		ApiURL:          providerInfo.APIURI,
		ProviderURL:     providerInfo.Domain,
		TosURL:          providerInfo.TosURL,
		ApiVersion:      apiVersion,
		AskForDonations: providerInfo.AskForDonations,
		Auth:            "annon",
		DonateURL:       providerInfo.DonateURL,
		HelpURL:         providerInfo.InfoURL,
		Provider:        name,
		CaCert:          string(caCert),
		GeolocationURL:  "",
	}

	return &providerOpts
}

func appendOnDiskProviders(providers *Providers) *Providers {
	provider_files, err := filepath.Glob(filepath.Join(config.Path, "provider_*.json"))
	if err != nil {
		log.Debug().
			Err(err).
			Msg("unable to locate on-disk providers JSON files")
		return providers
	}

	for _, f := range provider_files {
		data, err := os.ReadFile(f)
		if err != nil {
			log.Debug().
				Err(err).
				Msg("unable to read file")
			return providers
		}
		opts := &bitmask.ProviderOpts{}
		if err := json.Unmarshal(data, opts); err != nil {
			log.Debug().
				Err(err).
				Msg("unable to unmarshal provider JSON")
			return providers
		}
		exists := false
		for _, d := range providers.Data {
			if d.Provider == opts.Provider {
				exists = true
				break
			}
		}
		if !exists {
			providers.Data = append(providers.Data, *opts)
		}
	}
	return providers
}

func providerAlreadyExists(providers *Providers, provider *bitmask.ProviderOpts) bool {
	for _, p := range providers.Data {
		if p.Provider == provider.Provider {
			return true
		}
	}
	return false
}
