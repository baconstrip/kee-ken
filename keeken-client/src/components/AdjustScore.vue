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
    <div class="modal-box" role="document" style="max-width: 95vw">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Adjust Score for {{ player }}</h5>
        </div>
        <div class="modal-body">
            <label for="scoreAdjustment">Score Adjustment Amount:</label>
            <input type="number" id="scoreAdjustment" name="scoreAdjustment" v-model="amountField" />
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-info" @click="sendAdjustScore()">Submit</button>
          <button type="button" class="btn btn-secondary" @click="(adjustScoreModal! as any).close()">Close</button>
        </div>
      </div>
    </div>
  </dialog>
</template>