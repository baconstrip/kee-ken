<script setup lang="ts">

import { ref, useTemplateRef } from 'vue';
import eventBus from '../eventbus';

const player = ref<string | null>(null);

const adjustScoreModal = useTemplateRef<HTMLDivElement>('adjustScoreModal');
const amountField = ref<number | null>(null);

const sendAdjustScore = () => {
    const amountStr = amountField.value?.toString() || "";
    const amount = parseInt(amountStr, 10);
    if (isNaN(amount)) {
        alert("Please enter a valid number for the score adjustment.");
        return;
    }

    eventBus.emit("adjustScore", { playerName: player.value!, amount: amount});
    (adjustScoreModal! as any).value.close();
};

const reset = () => {
    amountField.value = null;
};

const open = (ply: string) => {
    reset();
    player.value = ply;
    (adjustScoreModal! as any).value.showModal();
};

defineExpose({
    open,
});
</script>

<template>
  <dialog class="modal" id="adjustScoreModal" tabindex="-1" role="dialog" data-backdrop="static" ref="adjustScoreModal">
    <div class="modal-box border-4 border-secondary" role="document">
      <div class="flex flex-col gap-4">
        <div class="modal-header">
          <div class="border-b-2 border-primary text-3xl p-3">Adjust Score for {{ player }}</div>
        </div>
        <div class="my-4 text-center">
            <label for="scoreAdjustment" class="input text-lg w-6/7"><span class="label">Score Adjustment Amount:</span>
              <input type="number" class="input-lg" id="scoreAdjustment" name="scoreAdjustment" v-model="amountField" />
            </label>
        </div>
        <div class="flex flex-row justify-end gap-4 mt-4">
          <button type="button" class="btn btn-info" @click="sendAdjustScore()">Submit</button>
          <button type="button" class="btn btn-secondary" @click="(adjustScoreModal! as any).close()">Close</button>
        </div>
      </div>
    </div>
  </dialog>
</template>