new Vue({
  el: '#join-app',

  data: {
  },

  methods: {
    clientJoin: function () {
      window.location.href = window.location + "/client"
    },
    hostJoin: function () {
      window.location.href = window.location + "/host"
    },
    editorJoin: function () {
      window.location.href = window.location + "/editor"
    },
  },
  template: `<div class="row text-center">
      <div class="col">
        <button type="button" class="btn btn-large btn-info" id="join-button" @click="clientJoin()">Launch Client</button>
        <button type="button" class="btn btn-large btn-disabled" id="join-button">Spectate</button> 
        <button type="button" class="btn btn-large btn-danger ml-5" id="join-button" @click="hostJoin()">Become Host</button>
        <button type="button" class="btn btn-large btn-primary ml-5" id="join-button" @click="editorJoin()">Launch Editor</button>
      </div>
    </div>
    `,
});

