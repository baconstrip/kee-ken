progressTemplate = `<div class="progress" style="height: 20px;">
  <div v-bind:class="classes" role="progressbar" v-bind:style="style"></div>
</div>`

Vue.component('progressbar', {
    props: ['begin', 'duration', 'color'],
    data: function () {
        return {
            start: 0,
            style: {
                width: "100%",
                "transition-duration": "0.1s",
            },
        };
    },
    template: progressTemplate,
    watch: {
        begin: function() {
            console.log("began countdown with duration: " + this.duration);
            
            this.start = new Date();
            var baseVue = this;
            var timer = setInterval(shrink, 50);
            this.timer = timer;
            this.$emit("countdownStart");
            function shrink() {
                var delta = new Date() - baseVue.start;
                if (delta > baseVue.duration) {
                    clearInterval(timer);
                    baseVue.style["width"] = 0;
                    return;
                }
                
                var frac = 1.0 - (delta / baseVue.duration) ;
                baseVue.style['width'] = frac * 100 + "%";
            }
        },
    },
    computed: {
        classes: function () {
            return ['progress-bar', 'bg-'+this.color]
        },
    },
});
