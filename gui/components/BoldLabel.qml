import QtQuick
import QtQuick.Controls
import "../themes/themes.js" as Theme
import "../themes"

Label {
    color: "black"

    font {
        pixelSize: Theme.fontSize * 1.5
        family: boldFontMonserrat.name
        bold: true
    }

    text: parent.text
    Accessible.name: text
    Accessible.role: Accessible.StaticText
}
