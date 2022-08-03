authTemplate = `<div class="container">
  <alertbox v-bind:message="serverError" v-if="serverError">
  </alertbox>
  <form>
    <div class="form-group">
      <label for="name-field">Name</label>
      <input type="name" class="form-control" id="name-field" placeholder="Pick a name">
    </div>
    <div class="form-group" v-if="host">
      <label for="server-passcode">Server Passcode</label>
      <input type="server-passcode" class="form-control" id="server-passcode" placeholder="Server Passcode">
      <small class="form-text text-muted">
        This was the passcode you set the server up with.
      </small>
    </div>
    <div class="form-group" v-else>
      <label for="passcode">Passcode</label>
      <input type="passcode" class="form-control" id="passcode" placeholder="Some passcode">
      <small class="form-text text-muted">
        Please do not use a real password here. This is just a passcode for the game, and may be shared with the server owner.
      </small>
    </div>
    <button type="button" class="btn btn-success" @click="sendAuth()">Join Up!</button>
  </form>
</div>
`

Vue.component('auth-window', {
  props: ['host'],
  template: authTemplate,
  data: function () {
    return {
      serverError: "",
    }
  },
  methods: {
    sendAuth: function () {
      baseVue = this;
      $.ajax({
        type: "POST",
        url: "http://" + window.location.host + "/auth",
        data: {
          Name: $("#name-field").val(),
          Host: baseVue.host == true,
          Passcode: $("#passcode").val(),
          ServerPasscode: $("#server-passcode").val(),
        },
        success: function () {
          console.log("Success");
          baseVue.$emit("auth-ready");
        },
        error: function (e) {
          msg = JSON.parse(e.responseText);
          baseVue.serverError = msg.Error;
        },
      });
    },
  },
});
