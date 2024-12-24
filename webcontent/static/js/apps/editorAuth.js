authTemplate = `<div class="container">
  <alertbox v-bind:message="serverError" v-if="serverError">
  </alertbox>
  <form class="login-form">
    <div class="form-group">
      <label for="name-field">Name</label>
      <input type="name" class="form-control" id="name-field" placeholder="Pick a name">
    </div>
    <div class="form-group">
      <label for="server-passcode">Server Passcode</label>
      <input type="server-passcode" class="form-control" id="server-passcode" placeholder="Server Passcode">
      <small class="form-text text-muted">
        This was the passcode you set the server up with.
      </small>
    </div>
    <button type="button" class="btn btn-success" @click="sendAuth()">Enter editor</button>
  </form>
</div>
`

Vue.component('editor-auth-window', {
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
          ServerPasscode: $("#server-passcode").val(),
          Passcode: "",
          Host: false,
          Editor: true,
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
