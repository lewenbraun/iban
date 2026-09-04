import QtQuick
import Quickshell
import Quickshell.Io

Scope {
    id: root

    property string currentState: "idle"
    property string sourceState: "idle"
    property bool blinkLit: true
    property var settings: ({
        "indicator": {
            "size_px": 14,
            "shape": "circle",
            "points": [{
                "corner": "top-left",
                "x_px": 16,
                "y_px": 16
            }, {
                "corner": "bottom-right",
                "x_px": 16,
                "y_px": 16
            }]
        },
        "states": {
            "recording": {
                "color": "#c95f67",
                "blink": false,
                "blink_interval_ms": 500
            },
            "transcribing": {
                "color": "#e5ad4f",
                "blink": false,
                "blink_interval_ms": 500
            },
            "done": {
                "color": "#62b879",
                "duration_ms": 3000,
                "blink": false,
                "blink_interval_ms": 500
            }
        }
    })
    readonly property var indicator: objectValue(settings, "indicator", ({
    }))
    readonly property var stateSettings: objectValue(settings, "states", ({
    }))
    readonly property var activeState: objectValue(stateSettings, currentState, ({
    }))
    readonly property int size: positiveInt(indicator.size_px, 14)
    readonly property string shape: indicator.shape === "square" ? "square" : "circle"
    readonly property string color: stringValue(activeState.color, "#ffffff")
    readonly property bool blinkEnabled: activeState.blink === true
    readonly property int blinkInterval: positiveInt(activeState.blink_interval_ms, 500)
    readonly property int doneDuration: positiveInt(objectValue(stateSettings, "done", ({
    })).duration_ms, 3000)
    readonly property bool isVisible: currentState !== "idle"
    readonly property bool isLit: !blinkEnabled || blinkLit

    function objectValue(object, key, fallback) {
        if (object === null || typeof object !== "object")
            return fallback;

        const value = object[key];
        if (value === null || typeof value !== "object")
            return fallback;

        return value;
    }

    function stringValue(value, fallback) {
        return typeof value === "string" && value.length > 0 ? value : fallback;
    }

    function positiveInt(value, fallback) {
        const number = Number(value);
        if (!Number.isFinite(number) || number <= 0)
            return fallback;

        return Math.round(number);
    }

    function integer(value, fallback) {
        const number = Number(value);
        if (!Number.isFinite(number))
            return fallback;

        return Math.round(number);
    }

    function pointValue(point, key, fallback) {
        if (point === null || typeof point !== "object")
            return fallback;

        if (key === "corner") {
            const corner = point.corner;
            const valid = corner === "top-left" || corner === "top-right" || corner === "bottom-left" || corner === "bottom-right";
            return valid ? corner : fallback;
        }
        return integer(point[key], fallback);
    }

    function point(index) {
        const points = indicator.points;
        if (!Array.isArray(points) || index >= points.length)
            return {
        };

        return points[index];
    }

    function updateState() {
        const rawState = stateFile.text().trim();
        const nextState = rawState === "recording" || rawState === "transcribing" || rawState === "done" ? rawState : "idle";
        if (nextState === sourceState)
            return ;

        sourceState = nextState;
        currentState = nextState;
        blinkLit = true;
        if (nextState === "done") {
            doneTimer.restart();
            return ;
        }
        doneTimer.stop();
    }

    function loadSettings() {
        try {
            const parsed = JSON.parse(configFile.text());
            if (parsed !== null && typeof parsed === "object")
                settings = parsed;

        } catch (error) {
        }
    }

    FileView {
        id: stateFile

        path: "/tmp/eban/indicator"
        watchChanges: true
        printErrors: false
        onLoaded: root.updateState()
        onFileChanged: reload()
    }

    FileView {
        id: configFile

        path: Qt.resolvedUrl("./indicator.json")
        watchChanges: true
        blockLoading: true
        printErrors: false
        onLoaded: root.loadSettings()
        onFileChanged: reload()
    }

    Timer {
        id: blinkTimer

        interval: root.blinkInterval
        running: root.isVisible && root.blinkEnabled
        repeat: true
        onTriggered: root.blinkLit = !root.blinkLit
    }

    Timer {
        interval: 250
        running: true
        repeat: true
        onTriggered: stateFile.reload()
    }

    Timer {
        id: doneTimer

        interval: root.doneDuration
        onTriggered: root.currentState = "idle"
    }

    Variants {
        model: Quickshell.screens

        delegate: Component {
            IndicatorWindow {
                required property var modelData

                controller: root
                point: root.point(0)
                targetScreen: modelData
            }

        }

    }

    Variants {
        model: Quickshell.screens

        delegate: Component {
            IndicatorWindow {
                required property var modelData

                controller: root
                point: root.point(1)
                targetScreen: modelData
            }

        }

    }

}
