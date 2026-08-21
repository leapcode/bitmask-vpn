function Controller()
{
    console.log("Controller script called")
}

Controller.prototype.ComponentSelectionPageCallback = function()
{
    gui.clickButton(buttons.NextButton);
}

Controller.prototype.ReadyForInstallationPageCallback = function() {
    console.log("Control script being called")
    try {
        var page = gui.pageWidgetByObjectName("DynamicInstallForAllUsersCheckBoxForm");
        if(page) {
            console.log("Control script being called")
            var choice = page.installForAllCheckBox.checked ? "true" : "false";
            installer.setValue("installForAllUsers", choice);
        } 
    } catch(e) {
        console.log(e);
    }
}

Controller.prototype.FinishedPageCallback = function() {
    if (systemInfo.productType === "macos") {
        var targetDir = installer.value("TargetDir")
        var appName = installer.value("Name")
        var src = targetDir + "/uninstall.app"
        var dstDir = targetDir + "/" + appName + ".app/Contents/Resources"
        var dst = dstDir + "/uninstall.app"
        if (installer.fileExists(src)) {
            console.log("Relocating maintenance tool: " + src + " -> " + dst)
            installer.execute("/bin/mkdir", ["-p", dstDir])
            installer.execute("/bin/mv", [src, dst])
        }
    }
}