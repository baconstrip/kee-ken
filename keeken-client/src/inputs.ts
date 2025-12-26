import eventBus from "./eventbus";

function keypressHandler(e: KeyboardEvent) {
    if (e.code === "Space") {
        eventBus.emit("spacePress");
    }

    if (e.code === "KeyY") {
        eventBus.emit("markCorrect");
    }

    if (e.code === "KeyN") {
        eventBus.emit("markIncorrect");
    }
}

function touchStartHandler() {
    eventBus.emit("spacePress");
}

export function setListeners() {
    window.removeEventListener("keydown", keypressHandler);
    window.removeEventListener("touchstart", touchStartHandler);

    window.addEventListener("keydown", keypressHandler);
    window.addEventListener("touchstart", touchStartHandler);
};
