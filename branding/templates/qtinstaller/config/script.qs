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