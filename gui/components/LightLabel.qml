import QtQuick
import QtQuick.Controls
import "../themes/themes.js" as Theme

Text {
    font.pixelSize: Theme.fontSize - 2
    font.family: Theme.fontFamily
    color: Theme.fontColor
    width: parent.width * 0.80
    text: parent.text

    horizontalAlignment: Text.AlignHCenter
    verticalAlignment: Text.AlignVCenter
    wrapMode: Text.Wrap

    Accessible.role: Accessible.StaticText
    Accessible.name: text
}
