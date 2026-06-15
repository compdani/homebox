<template>
  <BaseModal v-model="open">
    <template #title>{{ title }}</template>
    <form @submit.prevent="confirm">
      <FormTextField
        v-model.number="quantity"
        type="number"
        :min="1"
        :max="max"
        label="Quantity"
        :autofocus="true"
      />
      <div class="modal-action">
        <button type="button" class="btn" @click="open = false">Cancel</button>
        <BaseButton type="submit">Confirm</BaseButton>
      </div>
    </form>
  </BaseModal>
</template>

<script setup lang="ts">
  const props = defineProps<{
    modelValue: boolean;
    title?: string;
    max?: number;
  }>();

  const emit = defineEmits<{
    "update:modelValue": [value: boolean];
    confirm: [quantity: number];
  }>();

  const open = useVModel(props, "modelValue");
  const quantity = ref(1);

  const max = computed(() => props.max);

  watch(open, value => {
    if (value) {
      quantity.value = 1;
    }
  });

  function confirm() {
    if (quantity.value < 1) {
      return;
    }
    if (max.value && quantity.value > max.value) {
      return;
    }
    emit("confirm", quantity.value);
    open.value = false;
  }
</script>
