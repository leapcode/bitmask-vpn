import QtQuick 2.5
import QtQuick.Controls 2.14
import "../themes/themes.js" as Theme
import "../themes"

Label {
    FontLoader {
        id: boldFont
        source: "qrc:/oxanium-bold.ttf"
    }

    font.pixelSize: Theme.fontSize * 1.55555
    //font.family: boldFont.name
    font.bold: true
    //color: Theme.fontColorDark
    color: "black"
    text: parent.text
    //wrapMode: Text.WordWrap
    Accessible.name: text
    Accessible.role: Accessible.StaticText
}
