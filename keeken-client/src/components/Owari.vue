<script setup lang="ts">
import { onBeforeUnmount, ref } from 'vue';
import Alert from './Alert.vue';
import eventBus from '../eventbus';

const { category, host, prompt, answers, bids, money } = defineProps<{
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
  <div
    class="text-center md:w-full max-w-full grid grid-col-3 border-4 border-secondary/40 bg-secondary-content/70 shadow-2xl p-4 gap-4 md:max-w-3xl md:mx-auto mx-4 rounded-xl">
    <div
      class="col-span-3 text-center text-3xl text-secondary font-bold tracking-wide uppercase p-4 border-b-1 border-secondary">
      OWARI</div>
    <div class="col-span-3 text-center w-full mt-4 mb-8">
      <span class="text-base-content/50">Category</span>&emsp;
      <span class="text-accent text-xl">{{ category["Name"] }}</span>
    </div>
    <Alert :message="message" v-if="message"></Alert>
    <form v-if="!submitted && !host && !answers" class="col-span-3 border-t-1 border-secondary pt-4">
      <div class="flex flex-col gap-6 items-center mt-3">
        <label class="label">Bid Amount
          <input type="number" id="bid-amount" placeholder="Enter Bid" v-model="bidField">
        </label>
        <button class="btn btn-success" type="button" @click="submit()">Submit Bid</button>
      </div>
    </form>
    <div v-if="(submitted || host) && !answers && !ansSubmitted && !prompt" class="col-span-3">
      <div class="text-lg  animate-pulse col-span-3 text-center w-full">Waiting for players to submit their
        bids...</div>
    </div>

    <div v-if="prompt" class="col-span-3 text-center text-xl mb-4 text-secondary font-bold">
      <span>{{ prompt.Question }}</span>
    </div>
    <form v-if="prompt && !ansSubmitted && !answers && !host" @submit="sendAnswer()"
      class="col-span-3 border-t-1 border-secondary pt-4">
      <div class="flex flex-col gap-6 items-center mt-3">
        <label class="label">Answer
          <input type="text" id="owari-answer" placeholder="Enter your answer here" v-model="answerField">
        </label>
        <button class="btn btn-success" type="submit" @click="sendAnswer()">Submit Answer</button>
      </div>
    </form>
    <div v-if="host && prompt" class="col-span-3 text-center">
      <div class="badge badge-soft bg-secondary/20 badge-secondary badge-xl text-3xl py-6">
        <p>{{ prompt.Answer }}</p>
      </div>
    </div>
    <div v-if="answers" class="col-span-3 text-center grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="text-lg font-bold col-span-3 border-b-1 border-secondary pb-4">Submitted Answers</div>
      <div class="col-span-3 md:col-span-1" v-for="(n, a) in answers">
        <div class="bg-secondary/20 my-6 py-3 rounded-3xl">
          <div class="flex flex-col">
            <div class="text-base-content text-2xl font-bold">{{ a }}</div>
            <div class="text-secondary text-3xl">{{ n }}</div>
            <div class="mt-2 text-mdt text-base-content/70">
              Bid: {{ bids[a] }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>