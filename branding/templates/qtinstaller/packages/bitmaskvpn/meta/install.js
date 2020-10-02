/****************************************************************************
**
** Copyright (C) 2020 LEAP
**
****************************************************************************/

function Component() {
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
        var argList = ["-a", "@TargetDir@/DemoLibVPN.app"];
        try {
            installer.execute("touch", ["/tmp/install-finished"]);
            installer.execute("open", argList);
        } catch(e) {
            console.log(e);
        }
    }
}

function postInstallWindows() {
    component.addElevatedOperation("Execute", "@TargetDir@/helper.exe", "install", "UNDOEXECUTE", "@TargetDir@/helper.exe", "remove");
    component.addElevatedOperation("Execute", "@TargetDir@/helper.exe", "start", "UNDOEXECUTE", "@TargetDir@/helper.exe", "stop");
    console.log("Adding shortcut entries");
    component.addElevatedOperation("Mkdir", "@StartMenuDir@");
    component.addElevatedOperation("CreateShortcut", "@TargetDir@/demolib-vpn.exe", "@StartMenuDir@/DemoLibVPN.lnk", "workingDirectory=@TargetDir@", "iconPath=@TargetDir@/icon.ico", "description=Start DemoLibVPN");

    // TODO I think this one is not being created because the path doesn't exist yet. We might want to do this by hooking on the installation finished signal instead.
    component.addElevatedOperation(
        "CreateShortcut",
        "@TargetDir@/Uninstall-DemoLibVPN.exe",
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
    component.addOperation("AppendFile", "/tmp/riseupvpn.log", "this is a test - written from the installer");
}
