<script setup lang="ts">
import { ref } from 'vue';
import Auth from '../components/Auth.vue';
import eventBus from '@/eventbus';
//import ShowSelect from './editorShowSelect.vue';
const joined = ref(false);
const ws = ref<WebSocket | null>(null);
const errorMessages = ref("");

const { board } = defineProps<{
    board: any,
}>();

const join = () => {
    //$("#connect-button").addClass("d-none");
    ws.value = new WebSocket('ws://' + window.location.host + '/ws/editor');
    ws.value.onopen = function (e) {
        joined.value = true;
        errorMessages.value = "";
    };
    ws.value.onmessage = function (e) {
        var msg = JSON.parse(e.data);
        console.log(msg);
        if (msg["Type"] == "BoardOverview") {
            board.value = msg["Data"];
        }
    };
    ws.value.onerror = function (e) {
     //   $("#join-button").removeClass("d-none");
        errorMessages.value = "Couldn't establish connection to the server.";
        return;
    };
};

eventBus.on("authReady", () => {
    join();
});
</script>

<template>
    <div id="app" class="container">
        <!-- FIX this <Auth v-if="!joined"></Auth>-->
        <ShowSelect v-if="joined"></ShowSelect>
    </div>
</template>