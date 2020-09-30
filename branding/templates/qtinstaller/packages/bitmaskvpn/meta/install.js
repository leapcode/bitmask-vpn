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

function postInstallWindows() {
    component.addOperation(
	"CreateShortcut",
	"@TargetDir@/README.txt",
	"@StartMenuDir@/README.lnk",
	"workingDirectory=@TargetDir@",
	"iconPath=%SystemRoot%/system32/SHELL32.dll",
	"iconId=2");
}

function postInstallOSX() {
    console.log("Post-installation for OSX");
    // TODO add UNDOEXECUTE for the uninstaller
    component.addElevatedOperation(
	"Execute", "{0}",
   	"@TargetDir@/post-install.py",
	"errormessage=There was an error during the post-installation script, things might be broken. Please report this error and attach the post-install.log file.");
}

function postInstallLinux() {
    console.log("Post-installation for GNU/Linux");
    console.log("Maybe you want to use your package manager instead?");
    component.addOperation("AppendFile", "/tmp/riseupvpn.log", "this is a test - written from the installer");
}
