import { route } from "../base";
import { Requests } from "~~/lib/requests";
import { COLLECTIONS, authGroupId, getPb } from "~~/lib/pocketbase/client";
import { mapLocation } from "~~/lib/pocketbase/mappers";
import { wrap } from "~~/lib/pocketbase/response";
import type { LocationCreate, LocationOut, LocationOutCount, LocationUpdate } from "../types/data-contracts";
import type { TreeItem } from "../types/data-contracts";

export type LocationsQuery = {
  filterChildren: boolean;
};

export type TreeQuery = {
  withItems: boolean;
};

export class LocationsApi {
  constructor(private http: Requests) {}

  getAll(q: LocationsQuery = { filterChildren: false }) {
    return wrap(async () => {
      const records = await getPb().collection(COLLECTIONS.locations).getFullList({ sort: "name" });
      const filtered = q.filterChildren ? records.filter(r => !r.parent) : records;
      return filtered.map(
        (r): LocationOutCount => ({
          id: r.id,
          name: r.name,
          description: r.description || "",
          itemCount: 0,
          createdAt: r.created,
          updatedAt: r.updated,
        })
      );
    });
  }

  getTree(_tq = { withItems: false }) {
    return this.http.get<TreeItem[]>({ url: route("/locations/tree") });
  }

  create(body: LocationCreate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.locations)
        .create({
          name: body.name,
          description: body.description,
          parent: body.parentId || "",
          group: authGroupId(),
        });
      return mapLocation(rec);
    });
  }

  get(id: string) {
    return wrap(async () => {
      const rec = await getPb().collection(COLLECTIONS.locations).getOne(id, { expand: "parent" });
      return mapLocation(rec);
    });
  }

  delete(id: string) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.locations).delete(id);
    });
  }

  labelURL(id: string) {
    return `/api/v1/locations/${id}/label.png`;
  }

  update(id: string, body: LocationUpdate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.locations)
        .update(id, {
          name: body.name,
          description: body.description,
          parent: body.parentId || "",
        });
      return mapLocation(rec);
    });
  }
}
