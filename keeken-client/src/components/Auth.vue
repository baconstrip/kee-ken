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
  <div class="h-full max-w-screen flex items-center">
    <div class="m-auto grid gap-4 md:mx-auto mx-4 md:max-w-[600px] border-4 border-primary/40 bg-primary-content/50 rounded-xl p-4 md:p-8 shadow-2xl">
      <div class="text-3xl font-bold text-center text-secondary mb-4">
        <span>
          <template v-if="host">Join as Host</template>
          <template v-else-if="spectator">Join as a Spectator</template>
          <template v-else>Join Game!</template>
        </span>
      </div>
      <Alert :message="serverError" v-if="serverError">
      </Alert>
      <form class="login-form grid gap-4">
        <label class="label text-base-content">Display Name
          <input type="name" class="form-control" id="name-field" placeholder="Pick a name" v-model="nameField"></input>
        </label>
        <div class="w-full border-primary/40 border-2 rounded-md p-4 flex flex-col" v-if="host">
          <label class="label">Server Passcode
          <input type="server-passcode" class="form-control" id="server-passcode" placeholder="Server Passcode"
            v-model="serverPasscodeField"></input>
          </label>
          <div class="mt-4 text-sm text-base-content/70">
            This is the password that the server was set up with.
          </div>
        </div>
        <div class="form-group" v-else>
          <label class="label text-base-content">Passcode
            <input type="passcode" class="form-control" id="passcode" placeholder="Some passcode" v-model="passcodeField">
            </input>
          </label>
          <div class="mt-4 text-sm text-base-content/70">
            Please do not use a real password here. This is just a passcode for the game, and may be shared with the
            server owner.
          </div>
        </div>
        <button type="button" class="btn btn-success" @click="sendAuth()">Join Up!</button>
      </form>
    </div>
  </div>
</template>