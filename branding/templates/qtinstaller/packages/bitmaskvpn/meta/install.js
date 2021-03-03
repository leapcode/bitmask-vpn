/****************************************************************************
**
** Copyright (C) 2020 LEAP Encryption Access Project
**
****************************************************************************/

function majorVersion(str)
{
    return parseInt(str.split(".", 1));
}

function cancelInstaller(message)
{
    installer.setDefaultPageVisible(QInstaller.Introduction, false);
    installer.setDefaultPageVisible(QInstaller.TargetDirectory, false);
    installer.setDefaultPageVisible(QInstaller.ComponentSelection, false);
    installer.setDefaultPageVisible(QInstaller.ReadyForInstallation, false);
    installer.setDefaultPageVisible(QInstaller.StartMenuSelection, false);
    installer.setDefaultPageVisible(QInstaller.PerformInstallation, false);
    installer.setDefaultPageVisible(QInstaller.LicenseCheck, false);

    var abortText = "<font color='red'>" + message +"</font>";
    installer.setValue("FinishedText", abortText);
}

function Component() {
    // Check whether OS is supported.
    // start installer with -v to see debug output

    console.log("OS: " + systemInfo.productType);
    console.log("Kernel: " + systemInfo.kernelType + "/" + systemInfo.kernelVersion);

    var validOs = false;

    if (systemInfo.kernelType === "winnt") {
        if (majorVersion(systemInfo.kernelVersion) >= 6)
            validOs = true;
    } else if (systemInfo.kernelType === "darwin") {
        if (majorVersion(systemInfo.kernelVersion) >= 11)
            validOs = true;
    } else {
        if (systemInfo.productType !== "ubuntu"
                || systemInfo.productVersion !== "20.04") {
            QMessageBox["warning"]("os.warning", "Installer",
                                   "Note that the binaries are only tested on Ubuntu 20.04",
                                   QMessageBox.Ok);
        }
        validOs = true;
    }

    if (!validOs) {
        cancelInstaller("Installation on " + systemInfo.prettyProductName + " is not supported");
        return;
    }
    console.log("CPU Architecture: " +  systemInfo.currentCpuArchitecture);

}

Component.prototype.createOperations = function ()
{
    if (systemInfo.productType === "osx") {
        preInstallOSX();
    }
    if (systemInfo.productType === "windows") {
	preInstallWindows();
    }
    // This will actually install the files
    component.createOperations();

    // And now our custom actions.
    // See https://doc.qt.io/qtinstallerframework/operations.html for reference
    //
    // We can also use this to register different components (different architecture for instance)
    // See https://doc.qt.io/qtinstallerframework/qt-installer-framework-systeminfo-packages-root-meta-installscript-qs.html

    if (systemInfo.productType === "windows") {
        postInstallWindows();
    } else if (systemInfo.productType === "osx") {
        uninstallOSX();
        postInstallOSX();
    } else {
        postInstallLinux();
    }
}

function preInstallWindows() {
    console.log("Pre-installation for Windows: check for running bitmask");
    component.addOperation(
	"Execute", "{1}", "powershell", "-NonInteractive", "-NoProfile", "-command", "try {Get-Process $BINNAME} catch { exit 1}",
	"errormessage=It seems that an old RiseupVPN client is running. Please exit the app and run this installer again.",
    );
    /* Remove-Service only introduced in PS 6.0 */
    component.addElevatedOperation(
	"Execute", "{0}", "powershell", "-NonInteractive", "-NoProfile", "-command",
	"try {Get-Service bitmask-helper-v2} catch {exit 0}; try {Stop-Service bitmask-helper-v2} catch {}; try {$$srv = Get-Service bitmask-helper-v2; if ($$srv.Status -eq 'Running') {exit 1} else {exit 0};} catch {exit 0}",
	"errormessage=It seems that bitmask-helper-v2 service is running, and we could not stop it. Please manually uninstall any previous RiseupVPN or CalyxVPN client and run this installer again.",
    );
}

function postInstallWindows() {
    // TODO - check if we're on Windows10 or older, and use the needed tap-windows installer accordingly.
    console.log("Installing OpenVPN tap driver");
    component.addElevatedOperation("Execute", "@TargetDir@/tap-windows.exe", "/S", "/SELECT_UTILITIES=1");  /* TODO uninstall */
    /* remove an existing service, if it is stopped. Remove-Service is only in PS>6, and sc.exe delete leaves some garbage on the registry, so let's use the binary itself */
    console.log("Removing any previously installer helper...");
    component.addElevatedOperation("Execute", "{0,1}", "@TargetDir@/helper.exe", "remove");
    console.log("Now trying to install latest helper");
    component.addElevatedOperation("Execute", "@TargetDir@/helper.exe", "install", "UNDOEXECUTE", "@TargetDir@/helper.exe", "remove");
    component.addElevatedOperation("Execute", "@TargetDir@/helper.exe", "start", "UNDOEXECUTE", "@TargetDir@/helper.exe", "stop");
    console.log("Adding shortcut entries/...");
    component.addElevatedOperation("Mkdir", "@StartMenuDir@");
    component.addElevatedOperation("CreateShortcut", "@TargetDir@/$BINNAME.exe", "@StartMenuDir@/$APPNAME.lnk", "workingDirectory=@TargetDir@", "iconPath=@TargetDir@/icon.ico", "description=Start $APPNAME");

    // TODO I think this one is not being created because the path doesn't exist yet. We might want to do this by hooking on the installation finished signal instead.
    component.addElevatedOperation(
        "CreateShortcut",
        "@TargetDir@/Uninstall-$APPNAME.exe",
        "@StartMenuDir@/Uninstall.lnk"
    );
}

function preInstallOSX() {
    console.log("Pre-installation for OSX: check for running bitmask");
    component.addOperation(
	"Execute", "{1}", "pgrep", "bitmask-vpn$$", /* $$$$ is escaped by python template: the old app binary was called bitmask-vpn */ 
	"errormessage=It seems that an old RiseupVPN client is running. Please exit the app and run this installer again.",
    );
    component.addOperation(
	"Execute", "{1}", "pgrep", "bitmask$$", /* $$$$ is escaped by python template: we don't want to catch bitmask app */
	"errormessage=It seems RiseupVPN, CalyxVPN or LibraryVPN are running. Please exit the app and run this installer again.",
        "UNDOEXECUTE", "{1}", "pgrep", "bitmask$$", /* $$$$ is escaped: we dont want bitmask app */
	"errormessage=It seems RiseupVPN, CalyxVPN or LibraryVPN are running. Please exit the app before trying to run the uninstaller again."
    );
}

function uninstallOSX() {
    console.log("Pre-installation for OSX: uninstall previous helpers");
    // TODO use installer filepath??
    component.addElevatedOperation(
	"Execute", "{0}",
   	"@TargetDir@/uninstall.py", "pre",
	"errormessage=There was an error during the pre-installation script, things might be broken. Please report this error and attach /tmp/bitmask-uninstall.log"
    );
}

function postInstallOSX() {
    console.log("Post-installation for OSX");
    component.addElevatedOperation(
	"Execute", "{0}",
   	"@TargetDir@/post-install.py",
	"errormessage=There was an error during the post-installation script, things might be broken. Please report this error and attach the post-install.log file.",
        "UNDOEXECUTE",
        "@TargetDir@/uninstall.py"
    );
}

function postInstallLinux() {
    console.log("Post-installation for GNU/Linux");
    console.log("Maybe you want to use your package manager instead?");
    component.addOperation("AppendFile", "/tmp/bitmask-installer.log", "this is a test - written from the installer");
}
