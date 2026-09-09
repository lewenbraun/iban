import QtQuick
import Quickshell
import Quickshell.Wayland

PanelWindow {
    id: window

    required property var controller
    required property var point
    required property var targetScreen
    readonly property string corner: controller.pointValue(point, "corner", "top-left")
    readonly property bool isTop: corner === "top-left" || corner === "top-right"
    readonly property bool isRight: corner === "top-right" || corner === "bottom-right"
    readonly property int xOffset: controller.pointValue(point, "x_px", 0)
    readonly property int yOffset: controller.pointValue(point, "y_px", 0)

    screen: targetScreen
    visible: controller.isVisible
    color: "transparent"
    focusable: false
    implicitWidth: controller.size
    implicitHeight: controller.size
    exclusionMode: ExclusionMode.Ignore
    WlrLayershell.namespace: "iban-indicator"
    WlrLayershell.layer: WlrLayer.Overlay
    WlrLayershell.keyboardFocus: WlrKeyboardFocus.None

    surfaceFormat {
        opaque: false
    }

    anchors {
        top: window.isTop
        bottom: !window.isTop
        left: !window.isRight
        right: window.isRight
    }

    margins {
        top: window.isTop ? window.yOffset : 0
        bottom: window.isTop ? 0 : window.yOffset
        left: window.isRight ? 0 : window.xOffset
        right: window.isRight ? window.xOffset : 0
    }

    Rectangle {
        anchors.fill: parent
        color: window.controller.color
        radius: window.controller.shape === "circle" ? width / 2 : 0
        visible: window.controller.isLit
    }

    mask: Region {
    }

}
