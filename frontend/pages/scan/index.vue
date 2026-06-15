<script setup lang="ts">
  import { resolveAssetScan, resolveScanPayload, type ScanEntity } from "~~/lib/scan/resolve";

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "Homebox | Scan",
  });

  type ScanMode = "choose" | "product-first" | "location-batch" | "product-remove";
  type ProductFirstStep = "product" | "location";
  type BatchStep = "location" | "items";
  type RemoveStep = "product" | "location";

  const api = useUserApi();
  const toast = useNotifier();

  const mode = ref<ScanMode>("choose");
  const productFirstStep = ref<ProductFirstStep>("product");
  const batchStep = ref<BatchStep>("location");

  const scannedProduct = ref<ScanEntity | null>(null);
  const scannedUniqueItem = ref<ScanEntity | null>(null);
  const scannedLocation = ref<ScanEntity | null>(null);
  const batchLocation = ref<ScanEntity | null>(null);
  const removeStep = ref<RemoveStep>("product");
  const scannedRemoveProduct = ref<ScanEntity | null>(null);
  const scannedRemoveLocation = ref<ScanEntity | null>(null);

  const pendingEntity = ref<ScanEntity | null>(null);
  const showQuantity = ref(false);
  const quantityTitle = ref("Enter quantity");
  const quantityMax = ref<number | undefined>(undefined);
  const scannerPaused = ref(false);

  const recentBatch = ref<{ name: string; quantity: number }[]>([]);

  const scannerHint = computed(() => {
    if (mode.value === "product-first") {
      return productFirstStep.value === "product"
        ? "Scan a product or unique item QR code"
        : "Scan the destination location QR code";
    }
    if (mode.value === "location-batch") {
      return batchStep.value === "location"
        ? "Scan the location QR code to start batch mode"
        : "Scan product or unique item QR codes";
    }
    if (mode.value === "product-remove") {
      return removeStep.value === "product" ? "Scan the product QR code to remove" : "Scan the location QR code";
    }
    return "";
  });

  function resetProductFirst() {
    scannedProduct.value = null;
    scannedUniqueItem.value = null;
    scannedLocation.value = null;
    productFirstStep.value = "product";
    scannerPaused.value = false;
  }

  function resetBatch() {
    batchLocation.value = null;
    batchStep.value = "location";
    recentBatch.value = [];
    scannerPaused.value = false;
  }

  function resetRemove() {
    scannedRemoveProduct.value = null;
    scannedRemoveLocation.value = null;
    removeStep.value = "product";
    scannerPaused.value = false;
  }

  function startProductFirst() {
    mode.value = "product-first";
    resetProductFirst();
  }

  function startLocationBatch() {
    mode.value = "location-batch";
    resetBatch();
  }

  function startProductRemove() {
    mode.value = "product-remove";
    resetRemove();
  }

  async function entityName(entity: ScanEntity): Promise<string> {
    if (entity.type === "product") {
      const { data } = await api.products.get(entity.id);
      return data?.name || entity.id;
    }
    if (entity.type === "location") {
      const { data } = await api.locations.get(entity.id);
      return data?.name || entity.id;
    }
    const { data } = await api.items.get(entity.id);
    return data?.name || entity.id;
  }

  async function handleScan(raw: string) {
    let entity = resolveScanPayload(raw);
    if (!entity) {
      toast.error("Unrecognized QR code");
      return;
    }

    if (entity.type === "asset") {
      entity = await resolveAssetScan(api, entity);
      if (!entity) {
        toast.error("Asset ID not found");
        return;
      }
    }

    if (mode.value === "product-first") {
      await handleProductFirstScan(entity);
      return;
    }

    if (mode.value === "location-batch") {
      await handleBatchScan(entity);
      return;
    }

    if (mode.value === "product-remove") {
      await handleRemoveScan(entity);
    }
  }

  async function handleProductFirstScan(entity: ScanEntity) {
    if (productFirstStep.value === "product") {
      if (entity.type === "product") {
        scannedProduct.value = entity;
        scannedUniqueItem.value = null;
        productFirstStep.value = "location";
        toast.success("Product scanned — now scan location");
        return;
      }
      if (entity.type === "item") {
        scannedUniqueItem.value = entity;
        scannedProduct.value = null;
        productFirstStep.value = "location";
        toast.success("Item scanned — now scan location");
        return;
      }
      toast.error("Scan a product or unique item first");
      return;
    }

    if (entity.type !== "location") {
      toast.error("Scan a location QR code");
      return;
    }

    scannedLocation.value = entity;
    pendingEntity.value = scannedProduct.value || scannedUniqueItem.value;
    quantityTitle.value = `Quantity for ${await entityName(pendingEntity.value!)}`;
    scannerPaused.value = true;
    showQuantity.value = true;
  }

  async function handleRemoveScan(entity: ScanEntity) {
    if (removeStep.value === "product") {
      if (entity.type !== "product") {
        toast.error("Scan a product QR code first");
        return;
      }
      scannedRemoveProduct.value = entity;
      removeStep.value = "location";
      toast.success("Product scanned — now scan location to remove from");
      return;
    }

    if (entity.type !== "location") {
      toast.error("Scan a location QR code");
      return;
    }

    scannedRemoveLocation.value = entity;

    const { data, error } = await api.items.getAll({
      products: [scannedRemoveProduct.value!.id],
      locations: [entity.id],
      pageSize: 1,
    });

    if (error || !data?.items.length) {
      toast.error("Product not at this location");
      resetRemove();
      return;
    }

    const placement = data.items[0];
    quantityMax.value = placement.quantity;
    quantityTitle.value = `Remove from ${await entityName(entity)} (max ${placement.quantity})`;
    scannerPaused.value = true;
    showQuantity.value = true;
  }

  async function handleBatchScan(entity: ScanEntity) {
    if (batchStep.value === "location") {
      if (entity.type !== "location") {
        toast.error("Scan a location QR code first");
        return;
      }
      batchLocation.value = entity;
      batchStep.value = "items";
      toast.success(`Batch mode: ${await entityName(entity)}`);
      return;
    }

    if (entity.type !== "product" && entity.type !== "item") {
      toast.error("Scan a product or unique item");
      return;
    }

    pendingEntity.value = entity;
    quantityTitle.value = `Quantity for ${await entityName(entity)}`;
    scannerPaused.value = true;
    showQuantity.value = true;
  }

  async function onQuantityConfirm(quantity: number) {
    if (mode.value === "product-remove") {
      await onRemoveQuantityConfirm(quantity);
      return;
    }

    const location = mode.value === "product-first" ? scannedLocation.value : batchLocation.value;
    const entity = pendingEntity.value;

    if (!location || !entity) {
      scannerPaused.value = false;
      return;
    }

    if (entity.type === "product") {
      const { data, error } = await api.items.place({
        productId: entity.id,
        locationId: location.id,
        quantity,
      });
      if (error) {
        toast.error("Failed to place product");
      } else {
        toast.success(data.created ? "Product added" : "Quantity updated");
        if (mode.value === "location-batch") {
          recentBatch.value.unshift({ name: await entityName(entity), quantity });
        }
      }
    } else if (entity.type === "item") {
      const { error } = await api.items.place({
        itemId: entity.id,
        locationId: location.id,
        quantity,
      });
      if (error) {
        toast.error("Failed to update item");
      } else {
        toast.success("Item updated");
        if (mode.value === "location-batch") {
          recentBatch.value.unshift({ name: await entityName(entity), quantity });
        }
      }
    }

    pendingEntity.value = null;
    quantityMax.value = undefined;
    scannerPaused.value = false;
    showQuantity.value = false;

    if (mode.value === "product-first") {
      resetProductFirst();
    }
  }

  async function onRemoveQuantityConfirm(quantity: number) {
    const product = scannedRemoveProduct.value;
    const location = scannedRemoveLocation.value;

    if (!product || !location) {
      scannerPaused.value = false;
      showQuantity.value = false;
      return;
    }

    const { data, error } = await api.items.unplace({
      productId: product.id,
      locationId: location.id,
      quantity,
    });

    if (error) {
      toast.error("Failed to remove product");
    } else if (data.removed) {
      toast.success("Product removed from location");
    } else {
      toast.success(`${quantity} removed (${data.quantity} remaining)`);
    }

    quantityMax.value = undefined;
    scannerPaused.value = false;
    showQuantity.value = false;
    resetRemove();
  }

  function finishBatch() {
    mode.value = "choose";
    resetBatch();
  }
</script>

<template>
  <BaseContainer>
    <BaseSectionHeader>Scan</BaseSectionHeader>

    <div v-if="mode === 'choose'" class="grid gap-4 max-w-xl mx-auto mt-8">
      <button class="btn btn-lg btn-primary" @click="startProductFirst">Scan product, then location</button>
      <button class="btn btn-lg btn-secondary" @click="startLocationBatch">Scan location, then batch items</button>
      <button class="btn btn-lg btn-outline btn-warning" @click="startProductRemove">
        Remove product from location
      </button>
    </div>

    <div v-else class="max-w-2xl mx-auto space-y-6 mt-6">
      <div class="flex items-center gap-2">
        <button class="btn btn-sm" @click="mode = 'choose'">Back</button>
        <span class="text-sm opacity-70">
          {{
            mode === "product-first"
              ? productFirstStep === "product"
                ? "Step 1: Product or unique item"
                : "Step 2: Location"
              : mode === "product-remove"
                ? removeStep === "product"
                  ? "Step 1: Product"
                  : "Step 2: Location"
                : batchStep === "location"
                  ? "Step 1: Location"
                  : "Step 2: Batch scan items"
          }}
        </span>
        <button
          v-if="mode === 'location-batch' && batchStep === 'items'"
          class="btn btn-sm btn-accent ml-auto"
          @click="finishBatch"
        >
          Done
        </button>
      </div>

      <ScanCameraScanner :hint="scannerHint" :paused="scannerPaused" @scan="handleScan" />

      <div v-if="mode === 'location-batch' && recentBatch.length" class="card bg-base-100">
        <div class="card-body">
          <h3 class="font-semibold">This batch</h3>
          <ul class="text-sm space-y-1">
            <li v-for="(entry, index) in recentBatch" :key="index">{{ entry.name }} × {{ entry.quantity }}</li>
          </ul>
        </div>
      </div>
    </div>

    <ScanQuantityPrompt v-model="showQuantity" :title="quantityTitle" :max="quantityMax" @confirm="onQuantityConfirm" />
  </BaseContainer>
</template>
