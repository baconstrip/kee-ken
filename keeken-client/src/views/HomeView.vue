<script setup lang="ts">

</script>

<template>
  <main>
    <div class="row text-center">
      <div class="col">
        <RouterLink to="/player" type="button" class="btn btn-large btn-info" id="join-button">Launch Client</RouterLink>
        <RouterLink to="/host" type="button" class="btn btn-large btn-danger ml-5" id="join-button">Become Host</RouterLink>
        <RouterLink to="/editor" type="button" class="btn btn-large btn-primary ml-5" id="join-button">Launch Editor</RouterLink>
        <RouterLink to="" type="button" class="btn btn-large btn-disabled" id="join-button">Spectate</RouterLink> 
      </div>
    </div>
  </main>
</template>
