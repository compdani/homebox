<script setup lang="ts">
  import { useRegisterSW } from "virtual:pwa-register/vue";

  const { needRefresh, updateServiceWorker } = useRegisterSW();

  const dismissed = ref(false);

  const visible = computed(() => needRefresh.value && !dismissed.value);

  async function reload() {
    await updateServiceWorker(true);
  }

  function dismiss() {
    dismissed.value = true;
  }
</script>

<template>
  <div
    v-if="visible"
    class="fixed bottom-4 left-4 right-4 z-50 mx-auto flex max-w-lg items-center justify-between gap-4 rounded-lg bg-base-100 p-4 shadow-lg ring-1 ring-base-300"
    role="alert"
  >
    <p class="text-sm">A new version of Homebox is available.</p>
    <div class="flex shrink-0 gap-2">
      <button class="btn btn-ghost btn-sm" type="button" @click="dismiss">Dismiss</button>
      <button class="btn btn-primary btn-sm" type="button" @click="reload">Reload</button>
    </div>
  </div>
</template>
