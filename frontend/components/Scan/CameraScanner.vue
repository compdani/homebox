<template>
  <div class="space-y-4">
    <div v-if="cameraError" class="alert alert-error">{{ cameraError }}</div>

    <div class="relative overflow-hidden rounded-lg bg-black aspect-[4/3]">
      <video ref="videoRef" class="h-full w-full object-cover" playsinline muted />
      <div v-if="!active" class="absolute inset-0 flex items-center justify-center bg-base-300/80">
        <BaseButton @click="start">Start Camera</BaseButton>
      </div>
    </div>

    <p class="text-sm text-center opacity-70">{{ hint }}</p>
    <p v-if="lastScan" class="text-xs text-center opacity-60 break-all">Last scan: {{ lastScan }}</p>
  </div>
</template>

<script setup lang="ts">
  import { BrowserMultiFormatReader } from "@zxing/browser";

  const props = defineProps<{
    hint: string;
    paused?: boolean;
  }>();

  const emit = defineEmits<{
    scan: [value: string];
  }>();

  const videoRef = ref<HTMLVideoElement | null>(null);
  const active = ref(false);
  const cameraError = ref("");
  const lastScan = ref("");
  const reader = new BrowserMultiFormatReader();
  let controls: { stop: () => void } | null = null;
  let lastEmitted = "";
  let lastEmittedAt = 0;

  async function start() {
    if (props.paused) {
      return;
    }
    cameraError.value = "";
    if (!videoRef.value) {
      return;
    }
    try {
      controls = await reader.decodeFromVideoDevice(undefined, videoRef.value, result => {
        if (!result || props.paused) {
          return;
        }
        const text = result.getText();
        const now = Date.now();
        if (text === lastEmitted && now - lastEmittedAt < 2000) {
          return;
        }
        lastEmitted = text;
        lastEmittedAt = now;
        lastScan.value = text;
        emit("scan", text);
      });
      active.value = true;
    } catch (err) {
      cameraError.value = err instanceof Error ? err.message : "Unable to access camera";
    }
  }

  function stop() {
    controls?.stop();
    controls = null;
    active.value = false;
  }

  watch(
    () => props.paused,
    paused => {
      if (paused) {
        stop();
      } else if (!active.value) {
        start();
      }
    }
  );

  onMounted(() => {
    start();
  });

  onBeforeUnmount(() => {
    stop();
    BrowserMultiFormatReader.releaseAllStreams();
  });

  defineExpose({ start, stop });
</script>
