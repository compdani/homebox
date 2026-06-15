<script setup lang="ts">
  import type { ProductCreate } from "~~/lib/api/types/data-contracts";

  definePageMeta({
    middleware: ["auth"],
  });

  useHead({
    title: "Homebox | New Product",
  });

  const api = useUserApi();
  const toast = useNotifier();
  const loading = ref(false);

  const form = reactive<ProductCreate>({
    name: "",
    description: "",
    manufacturer: "",
    modelNumber: "",
  });

  async function create() {
    if (!form.name.trim()) {
      toast.error("Product name is required");
      return;
    }
    loading.value = true;
    const { data, error } = await api.products.create(form);
    loading.value = false;
    if (error) {
      toast.error("Couldn't create product");
      return;
    }
    toast.success("Product created");
    navigateTo(`/product/${data.id}`);
  }
</script>

<template>
  <BaseContainer>
    <BaseSectionHeader>New Product</BaseSectionHeader>
    <form class="max-w-2xl mx-auto my-5 space-y-4" @submit.prevent="create">
      <FormTextField v-model="form.name" :autofocus="true" label="Product Name" />
      <FormTextArea v-model="form.description" label="Description" />
      <FormTextField v-model="form.manufacturer" label="Manufacturer" />
      <FormTextField v-model="form.modelNumber" label="Model Number" />
      <BaseButton type="submit" :loading="loading">Create Product</BaseButton>
    </form>
  </BaseContainer>
</template>
