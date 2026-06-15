<template>
  <BaseModal v-model="modal">
    <template #title> Create Product </template>
    <form @submit.prevent="create()">
      <FormTextField v-model="form.name" :trigger-focus="focused" :autofocus="true" label="Product Name" />
      <FormTextArea v-model="form.description" label="Description" />
      <FormTextField v-model="form.manufacturer" label="Manufacturer" />
      <FormTextField v-model="form.modelNumber" label="Model Number" />
      <div class="modal-action">
        <BaseButton type="submit" :loading="loading">Create</BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  const props = defineProps({
    modelValue: {
      type: Boolean,
      required: true,
    },
  });

  const modal = useVModel(props, "modelValue");
  const loading = ref(false);
  const focused = ref(false);
  const form = reactive({
    name: "",
    description: "",
    manufacturer: "",
    modelNumber: "",
  });

  whenever(
    () => modal.value,
    () => {
      focused.value = true;
    }
  );

  const api = useUserApi();
  const toast = useNotifier();

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
    form.name = "";
    form.description = "";
    form.manufacturer = "";
    form.modelNumber = "";
    modal.value = false;
    navigateTo(`/product/${data.id}`);
  }
</script>
