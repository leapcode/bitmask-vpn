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
        postInstallOSX();
    } else {
        postInstallLinux();
    }
}

Component.prototype.installationFinished = function()
{
    console.log("DEBUG: running installationFinished");
    if (installer.isInstaller() && installer.status == QInstaller.Success) {
        var argList = ["-a", "@TargetDir@/$APPNAME.app"];
        try {
            installer.execute("touch", ["/tmp/install-finished"]);
            installer.execute("open", argList);
        } catch(e) {
            console.log(e);
        }
    }
}

function postInstallWindows() {
    console.log("Installing OpenVPN tap driver");
    component.addElevatedOperation("Execute", "@TargetDir@/tap-windows.exe", "/S", "/SELECT_UTILITIES=1");  /* TODO uninstall? */
    console.log("Now trying to install our helper");
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
