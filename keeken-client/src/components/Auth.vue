<script setup lang="ts">
import { ref } from 'vue';
import axios from 'axios';
import Alert from './Alert.vue';

import eventBus from '../eventbus';

const { host, spectator } = defineProps<{
  host: boolean,
  spectator: boolean,
}>();

const serverError = ref<string>("");

const nameField = ref<string>("");
const passcodeField = ref<string>("");
const serverPasscodeField = ref<string>("");


const sendAuth = () => {
  axios.post("http://" + window.location.host + "/api/auth", {
    Name: nameField.value,
    Host: host == true,
    Passcode: passcodeField.value,
    ServerPasscode: serverPasscodeField.value,
    Editor: false,
    Spectate: host == false && spectator == true,
  }, {
    headers: {
      'Content-Type': 'application/x-www-form-urlencoded'
    }
  }).then(() => {
    console.log("Auth Success");
    eventBus.emit("authReady");
  }).catch((e) => {
    const msg = e.response.data;
    serverError.value = msg.Error;
  });
};
</script>

<template>
  <div class="container">
    <Alert v-bind:message="serverError" v-if="serverError">
    </Alert>
    <form class="login-form">
      <div class="form-group">
        <label for="name-field">Name</label>
        <input type="name" class="form-control" id="name-field" placeholder="Pick a name" v-model="nameField">
      </div>
      <div class="form-group" v-if="host">
        <label for="server-passcode" class="text-4xl">Server Passcode</label>
        <input type="server-passcode" class="form-control" id="server-passcode" placeholder="Server Passcode" v-model="serverPasscodeField">
        <small class="form-text text-muted">
          This was the passcode you set the server up with.
        </small>
      </div>
      <div class="form-group" v-else>
        <label for="passcode">Passcode</label>
        <input type="passcode" class="form-control" id="passcode" placeholder="Some passcode" v-model="passcodeField">
        <small class="form-text text-muted">
          Please do not use a real password here. This is just a passcode for the game, and may be shared with the
          server owner.
        </small>
      </div>
      <button type="button" class="btn btn-success" @click="sendAuth()">Join Up!</button>
    </form>
  </div>
</template>