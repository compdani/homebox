<template>
  <FormAutocomplete2 v-if="products" v-model="value" :items="products" display="name" label="Product">
    <template #display="{ item, selected, active }">
      <div class="flex w-full">
        <div>
          <div>{{ item.value.name }}</div>
          <div v-if="item.value.manufacturer || item.value.modelNumber" class="text-xs opacity-70">
            {{ [item.value.manufacturer, item.value.modelNumber].filter(Boolean).join(" · ") }}
          </div>
        </div>
        <span
          v-if="selected"
          :class="['absolute inset-y-0 right-0 flex items-center pr-4', active ? 'text-white' : 'text-primary']"
        >
          <MdiCheck class="h-5 w-5" aria-hidden="true" />
        </span>
      </div>
    </template>
  </FormAutocomplete2>
</template>

<script lang="ts" setup>
  import type { ProductOut } from "~~/lib/api/types/data-contracts";
  import MdiCheck from "~icons/mdi/check";

  type Props = {
    modelValue?: ProductOut | null;
  };

  const props = defineProps<Props>();
  const value = useVModel(props, "modelValue");

  const productStore = useProductStore();
  const products = computed(() => productStore.products);
</script>
