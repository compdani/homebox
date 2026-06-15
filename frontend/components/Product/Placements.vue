<script setup lang="ts">
  import type { ItemSummary } from "~~/lib/api/types/data-contracts";
  import { onServerEvent, ServerEvent } from "~~/composables/use-server-events";
  import MdiMapMarker from "~icons/mdi/map-marker";
  import MdiMinus from "~icons/mdi/minus";
  import MdiPlus from "~icons/mdi/plus";
  import MdiQrcodeScan from "~icons/mdi/qrcode-scan";

  const props = defineProps<{
    productId: string;
  }>();

  const api = useUserApi();
  const toast = useNotifier();

  const { data: placements, refresh } = useAsyncData(`product-placements-${props.productId}`, async () => {
      const { data, error } = await api.items.getAll({
        products: [props.productId],
        pageSize: 500,
        orderBy: "name",
      });
      if (error || !data) {
        return [] as ItemSummary[];
      }
      return [...data.items].sort((a, b) => {
        const aName = a.location?.name || "";
        const bName = b.location?.name || "";
        return aName.localeCompare(bName);
      });
  });

  onServerEvent(ServerEvent.ItemMutation, () => {
    refresh();
  });

  const totalQuantity = computed(() => {
    return (placements.value || []).reduce((sum, p) => sum + p.quantity, 0);
  });

  const locationCount = computed(() => placements.value?.length || 0);

  async function adjustQuantity(placement: ItemSummary, amount: number) {
    const newQuantity = placement.quantity + amount;
    if (newQuantity < 0) {
      toast.error("Quantity cannot be negative");
      return;
    }

    const resp = await api.items.patch(placement.id, {
      id: placement.id,
      quantity: newQuantity,
    });

    if (resp.error) {
      toast.error("Failed to adjust quantity");
      return;
    }

    placement.quantity = newQuantity;
  }
</script>

<template>
  <section class="mt-6">
    <BaseSectionHeader class="mb-4"> Inventory </BaseSectionHeader>

    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-4">
      <div class="stat bg-base-100 rounded-box border border-base-300">
        <div class="stat-title">Total quantity</div>
        <div class="stat-value text-primary">{{ totalQuantity }}</div>
        <div class="stat-desc">Across all locations</div>
      </div>
      <div class="stat bg-base-100 rounded-box border border-base-300">
        <div class="stat-title">Locations</div>
        <div class="stat-value text-primary">{{ locationCount }}</div>
        <div class="stat-desc">Where this product is stored</div>
      </div>
    </div>

    <BaseCard v-if="placements && placements.length > 0">
      <table class="table w-full">
        <thead>
          <tr>
            <th class="text-no-transform text-sm bg-neutral text-neutral-content">Location</th>
            <th class="text-no-transform text-sm bg-neutral text-neutral-content text-center w-40">Quantity</th>
            <th class="text-no-transform text-sm bg-neutral text-neutral-content text-right w-32">Updated</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="placement in placements"
            :key="placement.id"
            class="hover cursor-pointer group"
            @click="navigateTo(`/item/${placement.id}`)"
          >
            <td class="bg-base-100">
              <NuxtLink
                v-if="placement.location"
                class="flex items-center gap-2 hover:link font-medium"
                :to="`/location/${placement.location.id}`"
                @click.stop
              >
                <MdiMapMarker class="h-4 w-4 shrink-0 opacity-70" />
                {{ placement.location.name }}
              </NuxtLink>
              <span v-else class="opacity-60">No location</span>
            </td>
            <td class="bg-base-100 text-center" @click.stop>
              <div class="inline-flex items-center gap-2">
                <button
                  class="btn btn-circle btn-xs opacity-0 group-hover:opacity-100 transition-opacity"
                  @click="adjustQuantity(placement, -1)"
                >
                  <MdiMinus class="h-3 w-3" />
                </button>
                <span class="badge badge-primary badge-lg min-w-[2.5rem]">{{ placement.quantity }}</span>
                <button
                  class="btn btn-circle btn-xs opacity-0 group-hover:opacity-100 transition-opacity"
                  @click="adjustQuantity(placement, 1)"
                >
                  <MdiPlus class="h-3 w-3" />
                </button>
              </div>
            </td>
            <td class="bg-base-100 text-right text-sm opacity-70">
              <DateTime :date="placement.updatedAt" />
            </td>
          </tr>
        </tbody>
      </table>
    </BaseCard>

    <div v-else class="bg-base-100 rounded-box border border-base-300 p-8 text-center">
      <p class="text-lg opacity-80 mb-2">Not placed in any location yet</p>
      <p class="text-sm opacity-60 mb-4">Scan a location QR code to add this product to inventory.</p>
      <NuxtLink to="/scan" class="btn btn-primary btn-sm">
        <MdiQrcodeScan class="mr-1 h-4 w-4" />
        Open Scanner
      </NuxtLink>
    </div>
  </section>
</template>
