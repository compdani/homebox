import { route } from "../base";
import { parseDate } from "../base/base-api";
import { Requests } from "~~/lib/requests";
import { COLLECTIONS, authGroupId, getPb } from "~~/lib/pocketbase/client";
import { buildItemsFilter, itemsSort } from "~~/lib/pocketbase/filters";
import {
  itemCreateBody,
  itemUpdateBody,
  mapItem,
  mapItemSummary,
  mapMaintenance,
} from "~~/lib/pocketbase/mappers";
import { wrap } from "~~/lib/pocketbase/response";
import type { RecordModel } from "pocketbase";
import type {
  ItemAttachmentUpdate,
  ItemCreate,
  ItemOut,
  ItemPatch,
  ItemPath,
  ItemSummary,
  ItemUpdate,
  MaintenanceEntry,
  MaintenanceEntryCreate,
  MaintenanceEntryUpdate,
  MaintenanceLog,
  PlaceItemRequest,
  PlaceItemResult,
  UnplaceItemRequest,
  UnplaceItemResult,
} from "../types/data-contracts";
import type { AttachmentTypes, PaginationResult } from "../types/non-generated";

export type ItemsQuery = {
  orderBy?: string;
  includeArchived?: boolean;
  page?: number;
  pageSize?: number;
  locations?: string[];
  labels?: string[];
  parentIds?: string[];
  products?: string[];
  q?: string;
  fields?: string[];
};

const itemExpand = "labels,location,parent,product";
const listExpand = "labels,location,parent,product";

async function loadItemRecord(id: string) {
  const rec = await getPb().collection(COLLECTIONS.items).getOne(id, { expand: itemExpand });
  let fields: RecordModel[] = [];
  let attachments: RecordModel[] = [];
  try {
    [fields, attachments] = await Promise.all([
      getPb().collection(COLLECTIONS.itemFields).getFullList({ filter: `item = "${id}"` }),
      getPb().collection(COLLECTIONS.attachments).getFullList({ filter: `item = "${id}"` }),
    ]);
  } catch {
    // Fields and attachments are optional; the item record is still usable without them.
  }
  rec.expand = { ...(rec.expand || {}), fields, attachments };
  return mapItem(rec);
}

export class AttachmentsAPI {
  constructor(private http: Requests) {}

  add(id: string, file: File | Blob, filename: string, type: AttachmentTypes | null = null) {
    return wrap(async () => {
      const form = new FormData();
      form.append("file", file, filename);
      form.append("item", id);
      form.append("group", authGroupId());
      if (type) {
        form.append("type", type);
      }
      form.append("title", filename);
      await getPb().collection(COLLECTIONS.attachments).create(form);
      return loadItemRecord(id);
    });
  }

  delete(id: string, attachmentId: string) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.attachments).delete(attachmentId);
    });
  }

  update(id: string, attachmentId: string, data: ItemAttachmentUpdate) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.attachments).update(attachmentId, {
        primary: data.primary,
        title: data.title,
        type: data.type,
      });
      return loadItemRecord(id);
    });
  }

  url(attachmentId: string, filename: string) {
    return getPb().files.getURL({ id: attachmentId, collectionId: COLLECTIONS.attachments } as any, filename);
  }
}

export class FieldsAPI {
  getAll() {
    return wrap(async () => {
      const records = await getPb().collection(COLLECTIONS.itemFields).getFullList({ fields: "name" });
      const names = new Set<string>();
      records.forEach(r => names.add(r.name));
      return Array.from(names);
    });
  }

  getAllValues(field: string) {
    return wrap(async () => {
      const records = await getPb()
        .collection(COLLECTIONS.itemFields)
        .getFullList({ filter: `name = "${field}"` });
      return records.map(r => r.text_value || String(r.number_value ?? "")).filter(Boolean);
    });
  }
}

type MaintenanceEntryQuery = {
  scheduled?: boolean;
  completed?: boolean;
};

export class MaintenanceAPI {
  getLog(itemId: string, q: MaintenanceEntryQuery = {}) {
    return wrap(async () => {
      const records = await getPb()
        .collection(COLLECTIONS.maintenance)
        .getFullList({ filter: `item = "${itemId}"`, sort: "-scheduled_date" });
      const entries = records.map(mapMaintenance);
      const log: MaintenanceLog = { entries };
      if (q.scheduled) {
        log.entries = entries.filter(e => e.scheduledDate);
      }
      if (q.completed) {
        log.entries = entries.filter(e => e.completedDate);
      }
      return log;
    });
  }

  create(itemId: string, data: MaintenanceEntryCreate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.maintenance)
        .create({
          item: itemId,
          group: authGroupId(),
          name: data.name,
          description: data.description,
          cost: data.cost ? Number(data.cost) : 0,
          scheduled_date: data.scheduledDate,
          date: data.completedDate,
        });
      return mapMaintenance(rec);
    });
  }

  delete(itemId: string, entryId: string) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.maintenance).delete(entryId);
    });
  }

  update(itemId: string, entryId: string, data: MaintenanceEntryUpdate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.maintenance)
        .update(entryId, {
          name: data.name,
          description: data.description,
          cost: data.cost ? Number(data.cost) : 0,
          scheduled_date: data.scheduledDate,
          date: data.completedDate,
        });
      return mapMaintenance(rec);
    });
  }
}

export class ItemsApi {
  attachments: AttachmentsAPI;
  maintenance: MaintenanceAPI;
  fields: FieldsAPI;

  constructor(private http: Requests, private attachmentToken: string) {
    this.fields = new FieldsAPI();
    this.attachments = new AttachmentsAPI(http);
    this.maintenance = new MaintenanceAPI();
  }

  fullpath(id: string) {
    return this.http.get<ItemPath[]>({ url: route(`/items/${id}/path`) });
  }

  getAll(q: ItemsQuery = {}) {
    return wrap(async () => {
      const page = q.page && q.page > 0 ? q.page : 1;
      const pageSize = q.pageSize && q.pageSize > 0 ? q.pageSize : 50;
      const result = await getPb()
        .collection(COLLECTIONS.items)
        .getList(page, pageSize, {
          filter: buildItemsFilter(q),
          sort: itemsSort(q),
          expand: listExpand,
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

  create(item: ItemCreate) {
    return wrap(async () => {
      const rec = await getPb()
        .collection(COLLECTIONS.items)
        .create(itemCreateBody(item, authGroupId()));
      const full = await loadItemRecord(rec.id);
      return full;
    });
  }

  async get(id: string) {
    const payload = await wrap(async () => loadItemRecord(id));
    if (!payload.error && payload.data) {
      payload.data = parseDate(payload.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    }
    return payload;
  }

  delete(id: string) {
    return wrap(async () => {
      await getPb().collection(COLLECTIONS.items).delete(id);
    });
  }

  async update(id: string, item: ItemUpdate) {
    const payload = await wrap(async () => {
      await getPb().collection(COLLECTIONS.items).update(id, itemUpdateBody(item));
      return loadItemRecord(id);
    });
    if (!payload.error && payload.data) {
      payload.data = parseDate(payload.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    }
    return payload;
  }

  async patch(id: string, item: ItemPatch) {
    const payload = await wrap(async () => {
      await getPb().collection(COLLECTIONS.items).update(id, { quantity: item.quantity });
      return loadItemRecord(id);
    });
    if (!payload.error && payload.data) {
      payload.data = parseDate(payload.data, ["purchaseTime", "soldTime", "warrantyExpires"]);
    }
    return payload;
  }

  place(body: PlaceItemRequest) {
    return this.http.post<PlaceItemRequest, PlaceItemResult>({
      url: route("/items/place"),
      body,
    });
  }

  unplace(body: UnplaceItemRequest) {
    return this.http.post<UnplaceItemRequest, UnplaceItemResult>({
      url: route("/items/unplace"),
      body,
    });
  }

  labelURL(id: string) {
    return `/api/v1/items/${id}/label.png`;
  }

  import(file: File | Blob) {
    const formData = new FormData();
    formData.append("csv", file);
    return this.http.post<FormData, void>({
      url: route("/items/import"),
      data: formData,
    });
  }

  exportURL() {
    return route("/items/export");
  }

  authURL(path: string): string {
    return path;
  }
}
