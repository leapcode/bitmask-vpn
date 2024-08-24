package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

/* Outline
 * runs as root and setup the bitmask-helper privileged helper on macOS
 * needs to perform the following steps:
 *  1. check if running as root
 *  2. setup the plist file with the correct path to bitmask-helper
 *  3. install plist file in location
 *  4. while doing the above make sure that existing helper is not running and removed
 */
const (
	plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>WorkingDirectory</key>
	<string>{{ .Path }}</string>
	<key>StandardOutPath</key>
	<string>{{ .Path }}/helper/bitmask-helper.log</string>
	<key>StandardErrorPath</key>
	<string>{{ .Path }}/helper/bitmask-helper-err.log</string>
	<key>GroupName</key>
	<string>daemon</string>
	<key>RunAtLoad</key>
	<true/>
	<key>SessionCreate</key>
	<true/>
	<key>KeepAlive</key>
    <true/>
    <key>ThrottleInterval</key>
    <integer>5</integer>
    <key>Label</key>
    <string>{{ .Label }}</string>
    <key>Program</key>
    <string>{{ .Path }}/bitmask-helper</string>
</dict>
</plist>`

	helperName = "bitmask-helper"

	// -action flag values
	actionPostInstall = "post-install"
	actionUninstall   = "uninstall"

	// -stage flag values
	stagePre       = "preinstall"
	stageUninstall = "uninstall"
)

var (
	curdir = func() string {
		execPath, err := os.Executable()
		if err != nil {
			log.Printf("error getting executable path: %v", err)
			return ""
		}
		return filepath.Dir(execPath)
	}()

	// flags
	installerAction string
	installerStage  string
	appName         string

	plistPath          string
	launchdDaemonLabel string
)

func init() {
	const (
		action  = "action"
		stage   = "stage"
		appname = "appname"
	)
	var usageAction = fmt.Sprintf("the installer actions: %s", strings.Join([]string{actionPostInstall, actionUninstall}, ","))
	var usageStage = "the installer action stage: preinstall, uninstall"
	var usageAppName = "name of the application being installed this is used to form the app bundle name by appending .app to it"

	flag.StringVar(&installerAction, action, "", usageAction)
	flag.StringVar(&installerStage, stage, stageUninstall, usageStage)
	flag.StringVar(&appName, appname, "", usageAppName)

	flag.Parse()
}

func main() {
	if os.Getuid() != 0 {
		log.Fatal("not running as root")
	}
	if appName == "" || installerAction == "" {
		log.Fatal("-action and -appname flags cannot be empty")
	}

	plistPath = fmt.Sprintf("/Library/LaunchDaemons/se.leap.helper.%s.plist", appName)
	launchdDaemonLabel = fmt.Sprintf("se.leap.Helper.%s", appName)

	switch installerAction {
	case actionPostInstall:
		if err := setupLogFile(filepath.Join(curdir, "post-install.log")); err != nil {
			log.Fatal(err)
		}
		log.Println("running action: post-install")
		if appBundlePath() == "" {
			log.Fatal("could not find path to .app bundle")
		}
		err := postInstall()
		if err != nil {
			log.Fatal(err)
		}
	case actionUninstall:
		log.Println("running action: uninstall")
		uninstall(installerStage)
	default:
		log.Fatalf("unknown command supplied: %s", installerAction)
	}
}

func appBundlePath() string {
	path := filepath.Join(curdir, appName+".app")
	_, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("unable to find the app bundle path: %v", err)
			return ""
		}
	}
	return path
}

func setupLogFile(logFile string) error {
	f, err := os.Create(logFile)
	if err != nil {
		return err
	}
	w := io.MultiWriter(os.Stdout, f)
	log.SetOutput(w)
	return nil
}

func postInstall() error {
	if isHelperRunning() {
		if err := unloadHelperPlist(); err != nil {
			log.Println(err)
		}
	}

	log.Println("Changing ownership of 'bitmask-helper'")
	// change ownership of bitmask-helper to root
	if err := os.Chown(filepath.Join(appBundlePath(), helperName), 0, 0); err != nil {
		log.Println("error while changing ownership of 'bitmask-helper': ", err)
	}
	// copy launchd plist file to target location /Library/LaunchDaemons
	log.Println("Generate plist file for helper launchd daemon")
	plist, err := generatePlist()
	if err != nil {
		return err
	}
	log.Println(plist)
	log.Println("Writing plist content to file")
	fout, err := os.OpenFile(plistPath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	if n, err := io.WriteString(fout, plist); n < len(plist) || err != nil {
		return fmt.Errorf("failed writing the plist file: %s: %v", fout.Name(), err)
	}

	// load the plist file onto launchd
	log.Println("Loading plist file")
	if err := loadHelperPlist(plistPath); err != nil {
		log.Printf("error while loading launchd daemon: %s: %v\n", plistPath, err)
	}

	// change ownership of 'helper' dir
	log.Println("Changing ownership of 'helper' dir")
	if err := os.Chown(filepath.Join(appBundlePath(), "helper"), 0, 0); err != nil {
		log.Println("error while changing ownership of dir 'helper': ", err)
	}
	return nil
}

func uninstall(stage string) {
	switch stage {
	case stagePre, stageUninstall:
		if err := setupLogFile(filepath.Join("/tmp", fmt.Sprintf("bitmask-%s.log", stage))); err != nil {
			log.Fatal(err)
		}
		if appBundlePath() == "" {
			log.Fatal("could not find path to .app bundle")
		}
	default:
		log.Fatal("unknow stage for uninstall: ", stage)
	}

	if isHelperRunning() {
		if err := unloadHelperPlist(); err != nil {
			log.Println("error while unloading launchd daemon: ", err)
		}
	}

	if err := os.Remove(plistPath); err != nil {
		log.Println("error while removing helper plist: ", err)
	}
}

func isHelperRunning() bool {
	cmd := exec.Command("ps", "-ceAo", "command")
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return false
	}
	processes := strings.Split(string(out), "\n")
	for _, proc := range processes {
		if strings.TrimSpace(proc) == "bitmask-helper" {
			return true
		}
	}
	return false
}

func loadHelperPlist(plistPath string) error {
	cmd := exec.Command("launchctl", "load", plistPath)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func unloadHelperPlist() error {
	cmd := exec.Command("launchctl", "unload", plistPath)
	_, err := cmd.Output()
	if err != nil {
		return err
	}
	cmd = exec.Command("launchctl", "remove", launchdDaemonLabel)
	_, _ = cmd.Output()
	cmd = exec.Command("pkill", "-9", helperName)
	_, err = cmd.Output()
	return err
}

func generatePlist() (string, error) {

	appPath := struct {
		Path  string
		Label string
	}{
		Path:  appBundlePath(),
		Label: launchdDaemonLabel,
	}

	t, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return "", err
	}

	plistContent := &bytes.Buffer{}
	err = t.Execute(plistContent, appPath)
	if err != nil {
		return "", err
	}

	return plistContent.String(), nil
}
