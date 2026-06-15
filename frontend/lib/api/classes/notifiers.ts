import { route } from "../base";
import { Requests } from "~~/lib/requests";
import { COLLECTIONS, authGroupId, getPb } from "~~/lib/pocketbase/client";
import { mapNotifier } from "~~/lib/pocketbase/mappers";
import { wrap } from "~~/lib/pocketbase/response";
import type { NotifierCreate, NotifierOut, NotifierUpdate } from "../types/data-contracts";

export class NotifiersAPI {
  constructor(private http: Requests) {}

  getAll() {
    return wrap(async () => {
      const records = await getPb().collection(COLLECTIONS.notifiers).getFullList({ sort: "name" });
      return records.map(mapNotifier) as NotifierOut[];
    });
  }

  create(body: NotifierCreate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.notifiers)
        .create({
          name: body.name,
          url: body.url,
          is_active: body.isActive ?? true,
          group: authGroupId(),
          user: getPb().authStore.model?.id,
        });
      return mapNotifier(rec);
    });
  }

  update(id: string, body: NotifierUpdate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.notifiers)
        .update(id, {
          name: body.name,
          url: body.url === "" ? null : body.url,
          is_active: body.isActive,
        });
      return mapNotifier(rec);
    });
  }

  delete(id: string) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.notifiers).delete(id);
    });
  }

  test(id: string) {
    return this.http.post<{ id: string }, null>({ url: route(`/notifiers/test`), body: { id } });
  }
}
