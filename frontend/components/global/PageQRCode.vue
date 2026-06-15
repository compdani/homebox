<template>
  <div class="dropdown dropdown-left">
    <slot>
      <label tabindex="0" class="btn btn-circle btn-sm">
        <MdiQrcode />
      </label>
    </slot>
    <div tabindex="0" class="card compact dropdown-content shadow-lg bg-base-100 rounded-box w-64">
      <div class="card-body">
        <h2 class="text-center">Scan URL</h2>
        <img :src="getQRCodeUrl()" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { route } from "../../lib/api/base";
  import MdiQrcode from "~icons/mdi/qrcode";

  const props = defineProps<{
    data?: string;
  }>();

  function getQRCodeUrl(): string {
    const payload = props.data || window.location.href;
    return route(`/qrcode`, { data: encodeURIComponent(payload) });
  }
</script>
