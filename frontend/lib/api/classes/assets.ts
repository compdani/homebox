import { route } from "../base";
import { Requests } from "~~/lib/requests";
import { COLLECTIONS, getPb } from "~~/lib/pocketbase/client";
import { mapItemSummary } from "~~/lib/pocketbase/mappers";
import { wrap } from "~~/lib/pocketbase/response";
import type { ItemSummary } from "../types/data-contracts";
import type { PaginationResult } from "../types/non-generated";

export class AssetsApi {
  constructor(private http: Requests) {}

  async get(id: string, page = 1, pageSize = 50) {
    return wrap(async () => {
      const result = await getPb()
        .collection(COLLECTIONS.items)
        .getList(page, pageSize, {
          filter: `asset_id = ${Number(id)}`,
          expand: "labels,location,product",
        });
      const payload: PaginationResult<ItemSummary> = {
        items: result.items.map(mapItemSummary),
        page: result.page,
        pageSize: result.perPage,
        total: result.totalItems,
      };
      return payload;
    });
  }
}
