import "themes.js" as Theme

Item {
    property alias fontFoo: fontFooLoader.name
    readonly property color colourBlackground: "#efefef"

    // TODO use theme.background
    FontLoader {
        id: fontFooLoader
        source: "qrc:/resources/fonts/Oxanium-Bold.ttf"
    }
}
