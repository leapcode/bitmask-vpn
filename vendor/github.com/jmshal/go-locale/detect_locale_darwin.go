package go_locale

func DetectLocale() (string, error) {
	return getCommandOutput(
		"defaults",
		"read",
		"/Library/Preferences/.GlobalPreferences",
		"AppleLocale")
}
