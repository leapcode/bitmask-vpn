/*
**
** Copyright (C) 2020 LEAP Encryption Access Project.
** 
*/

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
function majorVersion(str)
{
    return parseInt(str.split(".", 1));
}

function Component()
{
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

    if ( systemInfo.kernelType === "winnt") {
        installer.componentByName("root.win_x86_64").setValue("Default", "true");
        installer.componentByName("root.win_x86_64").setValue("Virtual", "false");
    }
}

Component.prototype.createOperations = function() 
{
	component.createOperations()
	component.addElevatedOperation("Execute", "@TargetDir@/helper.exe", "install", "UNDOEXECUTE", "@TargetDir@/helper.exe", "remove");
	component.addElevatedOperation("Execute", "@TargetDir@/helper.exe", "start", "UNDOEXECUTE", "@TargetDir@/helper.exe", "stop");
        if (systemInfo.productType === "windows") {
             console.log("Adding shortcut entries");
             component.addElevatedOperation("Mkdir", "@StartMenuDir@");
             component.addElevatedOperation("CreateShortcut", "@TargetDir@/demolib-vpn.exe", "@StartMenuDir@/DemoLibVPN.lnk", "workingDirectory=@TargetDir@", "iconPath=%SystemRoot%/system32/SHELL32.dll", "iconId=2", "description=Start DemoLibVPN");
             component.addElevatedOperation("CreateShortcut", "@TargetDir@/maintenancetool.exe", "@StartMenuDir@/uninstall.lnk", "workingDirectory=@TargetDir@", "iconPath=%SystemRoot%/system32/SHELL32.dll", "iconId=2", "description=Uninstall application");
        }
}
