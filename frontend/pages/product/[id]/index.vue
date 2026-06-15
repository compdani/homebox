<script setup lang="ts">
  import { downloadAuthedFile } from "~~/lib/api/download";
  import MdiBarcode from "~icons/mdi/barcode";
  import MdiPencil from "~icons/mdi/pencil";
  import MdiDelete from "~icons/mdi/delete";
  import MdiDownload from "~icons/mdi/download";

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();
  const toast = useNotifier();
  const productId = computed(() => route.params.id as string);

  const { data: product } = useAsyncData(productId.value, async () => {
    const { data, error } = await api.products.get(productId.value);
    if (error) {
      toast.error("Failed to load product");
      navigateTo("/products");
      return;
    }
    return data;
  });

  const confirm = useConfirm();

  async function confirmDelete() {
    const { isCanceled } = await confirm.open("Delete this product? Linked placements will remain as items.");
    if (isCanceled) {
      return;
    }
    const { error } = await api.products.delete(productId.value);
    if (error) {
      toast.error("Failed to delete product");
      return;
    }
    toast.success("Product deleted");
    navigateTo("/products");
  }

  async function downloadLabel() {
    const ok = await downloadAuthedFile(api.products.labelURL(productId.value), "label.png");
    if (!ok) {
      toast.error("Failed to download label");
    }
  }

  const qrData = computed(() => (product.value ? `${window.location.origin}/product/${product.value.id}` : ""));
</script>

<template>
  <BaseContainer v-if="product">
    <div class="bg-base-100 rounded p-3">
      <header class="mb-2">
        <div class="flex flex-wrap items-end gap-2">
          <div class="avatar placeholder mb-auto">
            <div class="bg-neutral-focus text-neutral-content rounded-full w-12">
              <MdiBarcode class="h-7 w-7" />
            </div>
          </div>
          <div>
            <h1 class="text-2xl pb-1">{{ product.name }}</h1>
            <p v-if="product.manufacturer || product.modelNumber" class="text-sm opacity-70">
              {{ [product.manufacturer, product.modelNumber].filter(Boolean).join(" · ") }}
            </p>
          </div>
          <div class="ml-auto mt-2 flex flex-wrap items-center gap-2">
            <PageQRCode :data="qrData" />
            <BaseButton size="sm" @click="downloadLabel">
              <MdiDownload class="mr-1" />
              Label
            </BaseButton>
            <NuxtLink :to="`/product/${product.id}/edit`" class="btn btn-sm">
              <MdiPencil class="mr-1" />
              Edit
            </NuxtLink>
            <BaseButton class="btn-sm" @click="confirmDelete">
              <MdiDelete class="mr-1" />
              Delete
            </BaseButton>
          </div>
        </div>
      </header>
      <div class="divider my-0 mb-1"></div>
      <Markdown v-if="product.description" :source="product.description" />
    </div>

    <ProductPlacements :product-id="product.id" />
  </BaseContainer>
</template>
