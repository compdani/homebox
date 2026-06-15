<template>
  <BaseModal v-model="modal">
    <template #title> Add to Location </template>

    <div class="tabs tabs-boxed mb-4">
      <button type="button" class="tab" :class="{ 'tab-active': mode === 'product' }" @click="mode = 'product'">
        Product
      </button>
      <button type="button" class="tab" :class="{ 'tab-active': mode === 'unique' }" @click="mode = 'unique'">
        Unique Item
      </button>
    </div>

    <form v-if="mode === 'product'" @submit.prevent="placeProduct()">
      <ProductSelector v-model="productForm.product" />
      <LocationSelector v-model="productForm.location" />
      <FormTextField v-model.number="productForm.quantity" type="number" min="1" label="Quantity" />
      <div class="modal-action">
        <BaseButton :loading="loading" type="submit">Add Product</BaseButton>
      </div>
    </form>

    <form v-else @submit.prevent="createUnique()">
      <LocationSelector v-model="uniqueForm.location" />
      <FormTextField
        ref="nameInput"
        v-model="uniqueForm.name"
        :trigger-focus="focused"
        :autofocus="true"
        label="Item Name"
      />
      <FormTextArea v-model="uniqueForm.description" label="Item Description" />
      <FormTextField v-model.number="uniqueForm.quantity" type="number" min="1" label="Quantity" />
      <FormMultiselect v-model="uniqueForm.labels" label="Labels" :items="labels ?? []" />
      <div class="modal-action">
        <div class="flex justify-center">
          <BaseButton class="rounded-r-none" :loading="loading" type="submit">Create</BaseButton>
          <div class="dropdown dropdown-top">
            <label tabindex="0" class="btn rounded-l-none rounded-r-xl">
              <MdiChevronDown class="h-5 w-5" />
            </label>
            <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-64 right-0">
              <li>
                <button type="button" @click="createUnique(false)">Create and Add Another</button>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  import type { ItemCreate, LabelOut, LocationOut, ProductOut } from "~~/lib/api/types/data-contracts";
  import { useLabelStore } from "~~/stores/labels";
  import { useLocationStore } from "~~/stores/locations";
  import MdiChevronDown from "~icons/mdi/chevron-down";

  const props = defineProps({
    modelValue: {
      type: Boolean,
      required: true,
    },
  });

  const api = useUserApi();
  const toast = useNotifier();
  const mode = ref<"product" | "unique">("product");

  const locationsStore = useLocationStore();
  const locations = computed(() => locationsStore.allLocations);

  const labelStore = useLabelStore();
  const labels = computed(() => labelStore.labels);

  const route = useRoute();

  const labelId = computed(() => {
    if (route.fullPath.includes("/label/")) {
      return route.params.id;
    }
    return null;
  });

  const locationId = computed(() => {
    if (route.fullPath.includes("/location/")) {
      return route.params.id;
    }
    return null;
  });

  const nameInput = ref<HTMLInputElement | null>(null);
  const modal = useVModel(props, "modelValue");
  const loading = ref(false);
  const focused = ref(false);

  const productForm = reactive({
    product: null as ProductOut | null,
    location: locations.value && locations.value.length > 0 ? locations.value[0] : ({} as LocationOut),
    quantity: 1,
  });

  const uniqueForm = reactive({
    location: locations.value && locations.value.length > 0 ? locations.value[0] : ({} as LocationOut),
    name: "",
    description: "",
    quantity: 1,
    labels: [] as LabelOut[],
  });

  const { shift } = useMagicKeys();

  whenever(
    () => modal.value,
    () => {
      focused.value = true;

      if (locationId.value) {
        const found = locations.value.find(l => l.id === locationId.value);
        if (found) {
          productForm.location = found;
          uniqueForm.location = found;
        }
      }

      if (labelId.value) {
        uniqueForm.labels = labels.value.filter(l => l.id === labelId.value);
      }
    }
  );

  async function placeProduct() {
    if (!productForm.product?.id || !productForm.location?.id) {
      toast.error("Select a product and location");
      return;
    }
    if (productForm.quantity < 1) {
      toast.error("Quantity must be at least 1");
      return;
    }

    loading.value = true;
    const { data, error } = await api.items.place({
      productId: productForm.product.id,
      locationId: productForm.location.id,
      quantity: Math.max(1, Number(productForm.quantity) || 1),
    });
    loading.value = false;

    if (error) {
      toast.error("Couldn't add product to location");
      return;
    }

    toast.success(data.created ? "Product added to location" : "Quantity updated");
    modal.value = false;
    navigateTo(`/item/${data.id}`);
  }

  async function createUnique(close = true) {
    if (!uniqueForm.location?.id || !uniqueForm.name.trim()) {
      toast.error("Name and location are required");
      return;
    }
    if (uniqueForm.quantity < 1) {
      toast.error("Quantity must be at least 1");
      return;
    }

    if (shift.value) {
      close = false;
    }

    loading.value = true;
    const out: ItemCreate = {
      parentId: null,
      name: uniqueForm.name,
      description: uniqueForm.description,
      locationId: uniqueForm.location.id,
      labelIds: uniqueForm.labels.map(l => l.id),
      quantity: uniqueForm.quantity,
    };

    const { error, data } = await api.items.create(out);
    loading.value = false;

    if (error) {
      toast.error("Couldn't create item");
      return;
    }

    toast.success("Item created");
    uniqueForm.name = "";
    uniqueForm.description = "";
    uniqueForm.quantity = 1;
    focused.value = false;

    if (close) {
      modal.value = false;
      navigateTo(`/item/${data.id}`);
    }
  }
</script>
