import { route } from "../base";
import { Requests } from "~~/lib/requests";
import { COLLECTIONS, authGroupId, getPb } from "~~/lib/pocketbase/client";
import { mapLabel } from "~~/lib/pocketbase/mappers";
import { wrap } from "~~/lib/pocketbase/response";
import type { LabelCreate, LabelOut } from "../types/data-contracts";

export class LabelsApi {
  constructor(private http: Requests) {}

  getAll() {
    return wrap(async () => {
      const result = await getPb().collection(COLLECTIONS.labels).getFullList({ sort: "name" });
      return result.map(mapLabel) as LabelOut[];
    });
  }

  create(body: LabelCreate) {
    return wrap(async () => {
      const rec = await getPb().collection(COLLECTIONS.labels).create({
        ...body,
        group: authGroupId(),
      });
      return mapLabel(rec);
    });
  }

  get(id: string) {
    return wrap(async () => {
      const rec = await getPb().collection(COLLECTIONS.labels).getOne(id);
      return mapLabel(rec);
    });
  }

  delete(id: string) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.labels).delete(id);
    });
  }

  update(id: string, body: LabelCreate) {
    return wrap(async () => {
      const rec = await getPb().collection(COLLECTIONS.labels).update(id, body);
      return mapLabel(rec);
    });
  }
}
