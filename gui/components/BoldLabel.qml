import QtQuick 2.5
import QtQuick.Controls 2.14
import "../themes/themes.js" as Theme
import "../themes"

Label {
    color: "black"

    font {
        pixelSize: Theme.fontSize * 1.5
        family: boldFont.name
        bold: true
    }

    text: parent.text
    Accessible.name: text
    Accessible.role: Accessible.StaticText
}
