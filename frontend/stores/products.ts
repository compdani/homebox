import { defineStore } from "pinia";
import type { ProductOut } from "~~/lib/api/types/data-contracts";

export const useProductStore = defineStore("products", {
  state: () => ({
    allProducts: null as ProductOut[] | null,
    client: useUserApi(),
  }),
  getters: {
    products(state): ProductOut[] {
      if (state.allProducts === null) {
        this.client.products.getAll().then(result => {
          if (result.error) {
            console.error(result.error);
            return;
          }

          this.allProducts = result.data;
        });
      }
      return state.allProducts ?? [];
    },
  },
  actions: {
    async refresh() {
      const result = await this.client.products.getAll();
      if (result.error) {
        return result;
      }

      this.allProducts = result.data;
      return result;
    },
  },
});
