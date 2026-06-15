<script setup lang="ts">
  import type { ProductOut } from "~~/lib/api/types/data-contracts";
  import MdiBarcode from "~icons/mdi/barcode";

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "Homebox | Products",
  });

  const api = useUserApi();
  const query = ref("");

  const { data: products } = useAsyncData("products-list", async () => {
    const { data, error } = await api.products.getAll();
    if (error) {
      return [] as ProductOut[];
    }
    return data;
  });

  const filtered = computed(() => {
    const list = products.value || [];
    const q = query.value.trim().toLowerCase();
    if (!q) {
      return list;
    }
    return list.filter(
      p =>
        p.name.toLowerCase().includes(q) ||
        p.manufacturer.toLowerCase().includes(q) ||
        p.modelNumber.toLowerCase().includes(q)
    );
  });
</script>

<template>
  <BaseContainer>
    <div class="flex flex-wrap items-center gap-3 mb-6">
      <BaseSectionHeader class="mb-0">Products</BaseSectionHeader>
      <NuxtLink to="/product/new" class="btn btn-primary btn-sm ml-auto">New Product</NuxtLink>
    </div>

    <FormTextField v-model="query" label="Search products" placeholder="Search by name, manufacturer, or model" />

    <div class="grid gap-3 mt-6 sm:grid-cols-2 lg:grid-cols-3">
      <NuxtLink
        v-for="product in filtered"
        :key="product.id"
        :to="`/product/${product.id}`"
        class="card bg-base-100 shadow hover:shadow-md transition-shadow"
      >
        <div class="card-body">
          <div class="flex items-start gap-3">
            <div class="avatar placeholder">
              <div class="bg-neutral-focus text-neutral-content rounded-full w-10">
                <MdiBarcode class="h-5 w-5" />
              </div>
            </div>
            <div>
              <h2 class="card-title text-lg">{{ product.name }}</h2>
              <p v-if="product.manufacturer || product.modelNumber" class="text-sm opacity-70">
                {{ [product.manufacturer, product.modelNumber].filter(Boolean).join(" · ") }}
              </p>
            </div>
          </div>
        </div>
      </NuxtLink>
    </div>

    <p v-if="filtered.length === 0" class="text-center opacity-70 mt-10">No products found.</p>
  </BaseContainer>
</template>
