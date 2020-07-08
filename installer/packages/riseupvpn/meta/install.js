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

    console.log("Post installation. Checking platform...")
    if (systemInfo.productType === "windows") {
        console.log("Platform: windows");
        postInstallWindows();
    } else if (systemInfo.productType === "osx") {
        console.log("Platform: osx");
        postInstallOSX();
    } else {
        console.log("Platform: linux");
        postInstallLinux();
    }
}

function postInstallWindows() {
    component.addOperation("CreateShortcut",
                   "@TargetDir@/README.txt",
                   "@StartMenuDir@/README.lnk",
                   "workingDirectory=@TargetDir@",
                   "iconPath=%SystemRoot%/system32/SHELL32.dll",
                   "iconId=2");
}

function postInstallOSX() {
    console.log("TODO: should do osx post-installation");
}

function postInstallLinux() {
    console.log("TODO: should do linux post-installation");
    console.log("Maybe you want to use your package manager instead?");
    component.addOperation("AppendFile", "/tmp/riseupvpn.log", "this is a test - written from the installer");
}
