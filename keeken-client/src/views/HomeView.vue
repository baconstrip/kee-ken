<script setup lang="ts">
import NavCard from '@/components/NavCard.vue';


</script>

<template>
    <div class="mt-10"></div>
    <div>
      <div class="row text-center mx-6 md:mx-0">
        <div class="bg-white rounded-[6rem] w-fit px-6 mx-auto fade-out-radial">
          <img src="/kiken_writing.png" alt="KeeKen Logo" class="mx-auto my-5 md:h-80 h-40" />
        </div>
      </div>

      <div class="mt-8 mx-auto row text-center">
        <span class="text-5xl">Kee-ken Trivia System</span>
      </div>
      <div class="mt-2 mx-auto row text-center">
        <span class="text text-gray-300">An open-source trivia platform for hosting and playing quiz games.</span>
      </div>
      <div class="mt-8 mx-auto row text-center">
        <div role="alert" class="mx-auto w-[100%] max-w-[700px] text-center alert alert-info alert-soft justify-center hidden md:block">
          <div class="text-2xl mx-auto">Select an option below to start playing!</div>
        </div>
      </div>

    </div>
    <div class="mt-10"></div>
    <div class="row text-center">
      <div class="col">
        <div class="grid grid-cols-3 justify-center gap-4 content-center">
          <NavCard btntext="Launch Client" link="/player" color="primary">
            <template #icon>
              <fa-icon icon="fa-solid fa-gamepad"></fa-icon>
            </template>
            <template #title>
              <span class="text-2xl font-bold">Play Now</span>
            </template>
            <template #description>
              <p class="mt-2">Join the game as a player and compete against others!</p>
            </template>
          </NavCard>
          <NavCard btntext="Become Host" link="/host" color="secondary" v-if="$matches.md!.min">
            <template #icon>
              <fa-icon icon="fa-solid fa-user-tie"></fa-icon>
            </template>
            <template #title>
              <span class="text-2xl font-bold">Host a Game</span>
            </template>
            <template #description>
              <p class="mt-2">Host your own game of Kee-ken</p>
            </template>
          </NavCard>
          <!-- Disable editor for the time being -->
          <!-- <NavCard btntext="Launch Editor" link="/editor" color="error" v-if="$matches.lg!.min">
            <template #icon>
              <fa-icon icon="fa-solid fa-edit"></fa-icon>
            </template>
            <template #title>
              <span class="text-2xl font-bold">Question Editor</span>
            </template>
            <template #description>
              <p class="mt-2">Create and edit question sets for your games!</p>
            </template>
          </NavCard> -->
          <NavCard btntext="Spectate" link="/spectate" color="accent">
            <template #icon>
              <fa-icon icon="fa-solid fa-eye"></fa-icon>
            </template>
            <template #title>
              <span class="text-2xl font-bold">Watch Game</span>
            </template>
            <template #description>
              <p class="mt-2">Join the game as a spectator!</p>
            </template>
          </NavCard>
          <!-- <RouterLink to="/host" type="button" class="btn btn-large btn-error mr-5" id="join-button">
            Become Host
          </RouterLink>
          <RouterLink to="/editor" type="button" class="btn btn-large btn-primary mr-5" id="join-button">
            Launch Editor
          </RouterLink>
          <RouterLink to="/spectate" type="button" class="btn btn-large btn-secondary" id="join-button">
            Spectate
          </RouterLink> -->
        </div>
      </div>
    </div>
</template>

<style lang="postcss" scoped>
.fade-out-edges {
    /* Create the fade-out effect */
    -webkit-mask-image: linear-gradient(to right, transparent, black 20%, black 80%, transparent);
    mask-image: linear-gradient(to right, transparent, black 20%, black 80%, transparent);

    /* For top and bottom fade-out, add this */
    -webkit-mask-image: linear-gradient(to bottom, transparent, black 20%, black 80%, transparent);
    mask-image: linear-gradient(to bottom, transparent, black 20%, black 80%, transparent);

    -webkit-mask-composite: destination-in;
    mask-composite: intersect;
}

.fade-out-radial {
  position: relative;
  background-size: cover;

  /* Radial fade-out effect */
  -webkit-mask-image: radial-gradient(circle, black 90%, transparent);
  mask-image: radial-gradient(circle, black 60%, transparent);

  -webkit-mask-composite: destination-in;
  mask-composite: intersect;
}
</style>
