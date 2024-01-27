import QtQuick

QtObject {
    // property var theme;
    // here we expose any var contained in 
    // the js file. This object can be accessed as the global 
    // customTheme, since it's loaded in main.qml
    // TODO it'd be real nice if we can implement a fallback mechanism so that any value defaults to the general theme.
    property string bgColor: theme.bgColor
    property string fgColor: theme.fgColor
    property string iconOn: theme.iconOn
    property string iconOff: theme.iconOff
    property string iconConnecting: theme.iconConnecting
    property string iconSplash: theme.iconSplash
}
