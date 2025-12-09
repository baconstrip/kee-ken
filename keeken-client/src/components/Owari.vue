<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue';
import Alert from './Alert.vue';
import eventBus from '../eventbus';

const { category, host, prompt, answers, bids, money  } = defineProps<{
    category: any,
    host: boolean,
    prompt: any,
    answers: any,
    bids: any,
    money: number,
}>();

const submitted = ref(false);
const ansSubmitted = ref(false);
const message = ref<string | null>(null);

const bidField = ref<number | null>(null);
const answerField = ref<string | null>(null);

const submit = () => {
    var amount = bidField.value;
    if (!amount) {
        message.value = "Enter a bid amount";
        return;
    }
    if (amount < 0) {
        message.value = "Enter a positive value";
        return;
    }
    if (amount! > money) {
        message.value = "Enter a value less than the amount you have";
        return;
    }
    eventBus.emit("submitBid", amount!);
    submitted.value = true;
}

const sendAnswer = () => {
    eventBus.emit("submitAnswer", answerField.value || "");
    ansSubmitted.value = true;
}

const hideBidListener = () => {
    submitted.value = true;
};

eventBus.on("hideBid", hideBidListener);

onBeforeUnmount(() => {
    eventBus.off("hideBid", hideBidListener);
});
</script>

<template>
    <div class="text-center" id="owari-board">
        <h3>Category: {{ category["Name"] }}</h3>
        <Alert :message="message" v-if="message"></Alert>
        <form v-if="!submitted && !host && !answers">
            <div class="form-group">
                <label for="bid">Bid Amount</label> 
                <input type="number" id="bid-amount" placeholder="Enter Bid" v-model="bidField">
            </div>
            <button class="btn btn-success" type="button" @click="submit()">Submit Bid</button>
        </form>

        <form v-if="prompt && !ansSubmitted && !answers" @submit="sendAnswer()">
            <h2>Question: {{ prompt.Question }}</h2>
            <label v-if="!host" for="answer">Answer</label>
            <input v-if="!host" type="text" id="owari-answer" placeholder="Enter your answer here" v-model="answerField">
            <button v-if="!host" class="btn btn-danger" type="submit" @click="sendAnswer()">Submit Answer</button>
        </form>
        <h3 v-if="host && prompt"><i>Answer: {{ prompt.Answer }}</i></h3>
        <ul>
            <li v-for="(n, a) in answers" v-if="answers">
                <em>{{ a }}: </em>{{ n }} (Bid: {{ bids[a] }})
            </li>
        </ul>
    </div>
</template>