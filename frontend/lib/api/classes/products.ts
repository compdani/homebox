import type { ProductCreate, ProductOut, ProductUpdate } from "../types/data-contracts";
import { COLLECTIONS, authGroupId, getPb } from "~~/lib/pocketbase/client";
import { mapProduct } from "~~/lib/pocketbase/mappers";
import { wrap } from "~~/lib/pocketbase/response";

export class ProductsApi {
  getAll() {
    return wrap(async () => {
      const result = await getPb().collection(COLLECTIONS.products).getFullList({ sort: "name" });
      return result.map(mapProduct) as ProductOut[];
    });
  }

  create(body: ProductCreate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.products)
        .create({
          name: body.name,
          description: body.description || "",
          manufacturer: body.manufacturer || "",
          model_number: body.modelNumber || "",
          group: authGroupId(),
        });
      return mapProduct(rec);
    });
  }

  get(id: string) {
    return wrap(async () => {
      const rec = await getPb().collection(COLLECTIONS.products).getOne(id);
      return mapProduct(rec);
    });
  }

  delete(id: string) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.products).delete(id);
    });
  }

  update(id: string, body: ProductUpdate) {
    return wrap(async () => {
      const rec = await getPb().collection(COLLECTIONS.products).update(id, {
        name: body.name,
        description: body.description,
        manufacturer: body.manufacturer,
        model_number: body.modelNumber,
      });
      return mapProduct(rec);
    });
  }

  labelURL(id: string) {
    return `/api/v1/products/${id}/label.png`;
  }
}
