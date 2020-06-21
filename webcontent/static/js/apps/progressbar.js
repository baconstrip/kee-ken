progressTemplate = `<div class="progress" style="height: 20px;">
  <div v-bind:class="classes" role="progressbar" v-bind:style="style"></div>
</div>`

Vue.component('progressbar', {
    props: ['duration'],
    data: function () {
        return {
            start: 0,
            style: {
                width: "100%",
                "transition-duration": "0.1s",
            },
            color: 'danger',
        };
    },
    template: progressTemplate,
    methods: {
        beginCountdown: function(e) {
            console.log("began countdown with duration: " + this.duration);
            if (e == "answer") {
                this.color = "success";
            }
            else if (e == "buzz")  {
                this.color = "warning";
            }
            
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
    created: function () {
        EventBus.$on("beginCountdown", this.beginCountdown); 
    },
    beforeDestroy: function() {
        EventBus.$off("beginCountdown", this.beginCountdown);
    },
});
