EventBus = new Vue();

document.addEventListener("keypress", function (e) {
    if (event.key == " ") {
        EventBus.$emit("spacePress");
    }
});

document.addEventListener("touchstart", function (e) {
    EventBus.$emit("spacePress");
});
