<script setup lang="ts">
  import type { ProductUpdate } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  const route = useRoute();
  const api = useUserApi();
  const toast = useNotifier();
  const loading = ref(false);
  const productId = computed(() => route.params.id as string);

  const form = reactive<ProductUpdate>({
    id: productId.value,
    name: "",
    description: "",
    manufacturer: "",
    modelNumber: "",
  });

  onMounted(async () => {
    const { data, error } = await api.products.get(productId.value);
    if (error || !data) {
      toast.error("Failed to load product");
      navigateTo("/products");
      return;
    }
    form.name = data.name;
    form.description = data.description;
    form.manufacturer = data.manufacturer;
    form.modelNumber = data.modelNumber;
  });

  async function save() {
    loading.value = true;
    const { error } = await api.products.update(productId.value, form);
    loading.value = false;
    if (error) {
      toast.error("Failed to update product");
      return;
    }
    toast.success("Product updated");
    navigateTo(`/product/${productId.value}`);
  }
</script>

<template>
  <BaseContainer>
    <BaseSectionHeader>Edit Product</BaseSectionHeader>
    <form class="max-w-2xl mx-auto my-5 space-y-4" @submit.prevent="save">
      <FormTextField v-model="form.name" label="Product Name" />
      <FormTextArea v-model="form.description" label="Description" />
      <FormTextField v-model="form.manufacturer" label="Manufacturer" />
      <FormTextField v-model="form.modelNumber" label="Model Number" />
      <div class="flex gap-2">
        <BaseButton type="submit" :loading="loading">Save</BaseButton>
        <NuxtLink :to="`/product/${productId}`" class="btn">Cancel</NuxtLink>
      </div>
    </form>
  </BaseContainer>
</template>
