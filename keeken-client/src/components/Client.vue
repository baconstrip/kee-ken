<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';

import eventBus from '../eventbus';

import Gameboard from './Gameboard.vue';
import Gameheader from './Gameheader.vue';
import Alert from './Alert.vue';
import Owari from './Owari.vue';
import Question from './Question.vue';
import PlayerList from './PlayerList.vue';
import Auth from './Auth.vue';
import { useRoute } from 'vue-router';

const ws = ref<WebSocket | null>(null);
const joined = ref(false);
const errorMessages = ref('');
const gameErrors = ref('');
const board = ref<any>(null);
const hostPlayerName = ref("");
const question = ref('');
const answer = ref('');
const players = ref<any>(null);

const owari = ref<any>(null);
const owariPrompt = ref<any>(null);
const owariAnswers = ref<any>(null);
const owariBids = ref<any>(null);

const yourMoney = ref(-1);

const questionComponent = ref<string | null>(null);

const duration = ref(0);
const answeringPlayer = ref<string | null>(null);
const responsesClosed = ref(false);

const route = useRoute();

const host = computed(() => {
    return route.path.includes("host");
});

// ------------- WebSocket receivers ----------------
const wsMessageListener = (rawMessage: any) => {
    var msg = JSON.parse(rawMessage.data);
    console.log(msg);

    if (msg["Type"] == "BoardOverview") {
        board.value = msg["Data"];
    }
    if (msg["Type"] == "QuestionPrompt") {
        questionComponent.value = "question";
        question.value = msg["Data"].Question;
        answer.value = msg["Data"].Answer || "";
    }
    if (msg["Type"] == "UpdatePlayers") {
        players.value = msg["Data"].Plys;
    }
    if (msg["Type"] == "HostAdd") {
        hostPlayerName.value = msg["Data"].Name;
    }
    if (msg["Type"] == "OpenResponses") {
        answeringPlayer.value = null;
        duration.value = msg["Data"].Interval;
        eventBus.emit("beginCountdown", "buzz");
    }
    if (msg["Type"] == "PlayerAnswering") {
        duration.value = msg["Data"].Interval;
        answeringPlayer.value = msg["Data"].Name;
        eventBus.emit("beginCountdown", "answer");
    }
    if (msg["Type"] == "CloseResponses") {
        responsesClosed.value = true;
    }
    if (msg["Type"] == "HideQuestion") {
        eventBus.emit("hideQuestion");
        questionComponent.value = null;
        answeringPlayer.value = null;
        duration.value = 0;
        responsesClosed.value = false;
        question.value = '';
        answer.value = '';
    }
    if (msg["Type"] == "ServerError") {
        gameErrors.value = msg["Data"].Error;
        setTimeout(() => { gameErrors.value = '' }, 10000);
    }
    if (msg["Type"] == "BeginOwari") {
        owari.value = msg["Data"].Category;
        yourMoney.value = msg["Data"].Money;
    }
    if (msg["Type"] == "ShowOwariPrompt") {
        owariPrompt.value = msg["Data"].Prompt;
    }
    if (msg["Type"] == "ShowOwariResults") {
        owariBids.value = msg["Data"].Bids;
        owariAnswers.value = msg["Data"].Answers;
    }
    if (msg["Type"] == "ClearBoard") {
        errorMessages.value = '';
        gameErrors.value = '';
        board.value = null;
        owari.value = null;
        owariPrompt.value = null;
        owariAnswers.value = null;
        owariBids.value = null;
        question.value = '';
    }
};

const join = () => {
    ws.value = new WebSocket('ws://' + window.location.host + '/player_game');
    ws.value.onopen = function (e) {
        joined.value = true;
        errorMessages.value = "";
    };

    ws.value.onmessage = wsMessageListener;
    ws.value.onerror = function (e) {
        //       $("#join-button").removeClass("d-none");
        errorMessages.value = "Couldn't establish connection to the server.";
        return;
    };
};

// ------------- WebSocket senders ----------------
const sendWSMessage = (type: string, data: any) => {
    ws.value!.send(
        JSON.stringify({
            Type: type,
            Data: data,
        })
    );
    console.log({ Type: type, Data: data });
};

const sendSelect = (e: any) => {
    sendWSMessage("SelectQuestion", { ID: e.value });
};

const sendFinishReading = () => {
    sendWSMessage("FinishReading", {});
};

const sendBuzz = (e: any) => {
    sendWSMessage("AttemptAnswer", { ResponseTime: e });
};

const sendMarkAnswer = (correct: boolean) => {
    sendWSMessage("MarkAnswer", { Correct: correct == true });
};

const sendMoveOn = () => {
    sendWSMessage("MoveOn", {});
};

const sendNextRound = () => {
    sendWSMessage("NextRound", {});
};

const sendStartGame = () => {
    sendWSMessage("StartGame", {});
};

const sendBid = (e: any) => {
    sendWSMessage("EnterBid", { Money: Number(e) });
};

const sendOwariAnswer = (e: any) => {
    sendWSMessage("FreeformAnswer", { Message: e });
};

const sendCancelGame = () => {
    sendWSMessage("CancelGame", {});
};
// ------------------------------------------------


// ------------- Global page listeners ----------------
document.addEventListener("keypress", function (e) {
    if (e.key == " ") {
        eventBus.emit("spacePress");
    }
});

document.addEventListener("touchstart", function (e) {
    eventBus.emit("spacePress");
});
// ----------------------------------------------------

// ------------- EventBus listeners ----------------
eventBus.on("cancelGame", sendCancelGame);
eventBus.on("selectQuestion", sendSelect);
eventBus.on("nextRound", sendNextRound);
eventBus.on("markCorrect", () => (sendMarkAnswer(true)));
eventBus.on("markIncorrect", () => (sendMarkAnswer(false)));
eventBus.on("finishReading", sendFinishReading);
eventBus.on("moveOn", sendMoveOn);
eventBus.on("buzz", sendBuzz);
eventBus.on("submitBid", sendBid);
eventBus.on("submitAnswer", sendOwariAnswer);
eventBus.on("authReady", join);
// ----------------------------------------------------
</script>


<template>
    <div id="app" class="container">
        <div class="alert alert-warning" role="alert" v-if="errorMessages">
            <span v-html="errorMessages"></span>
        </div>
        <Gameheader v-if="ws" v-bind:hostName="hostPlayerName" v-bind:board="board" v-bind:host="host"
            v-bind:started="!!board">
        </Gameheader>
        <Gameboard v-if="board" v-bind:board="board" v-bind:host="host">
        </Gameboard>
        <Alert message="Waiting for the host to start the game..." v-if="!board && joined">
        </Alert>
        <button v-if="joined && host && !board" type="button" class="btn btn-success" @click="sendStartGame">Start
            Game</button>
        <Alert message="Connecting to the server..." v-if="ws == null && joined">
        </Alert>
        <component v-bind:is="questionComponent" v-bind:question="Question" v-if="questionComponent" :key="Question"
            v-bind:answer="answer" v-bind:host="host" v-bind:duration="duration"
            v-bind:answeringPlayer="answeringPlayer" v-bind:responsesClosed="responsesClosed">
        </component>
        <Owari v-if="Owari" v-bind:category="Owari" v-bind:prompt="owariPrompt" v-bind:answers="owariAnswers"
            v-bind:bids="owariBids" v-bind:money="yourMoney" v-bind:host="host">
        </Owari>
        <Auth v-bind:host="host" v-if="!joined">
        </Auth>
        <Alert v-bind:message="gameErrors" v-if="gameErrors">
        </Alert>
        <PlayerList v-if="players" v-bind:players="players">
        </PlayerList>
    </div>
</template>
