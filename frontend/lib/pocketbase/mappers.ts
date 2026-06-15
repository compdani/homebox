import type { RecordModel } from "pocketbase";
import type {
  Group,
  ItemAttachment,
  ItemField,
  ItemOut,
  ItemSummary,
  LabelOut,
  LabelSummary,
  LocationOut,
  LocationSummary,
  MaintenanceEntry,
  NotifierOut,
  ProductOut,
  ProductSummary,
  UserOut,
} from "../api/types/data-contracts";
import { COLLECTIONS, getPb } from "./client";
import { recordDates, toDate } from "./response";

export function mapUser(rec: RecordModel, expand?: Record<string, any>): UserOut {
  const group = expand?.group || rec.expand?.group;
  return {
    id: rec.id,
    name: rec.name,
    email: rec.email,
    groupId: rec.group,
    groupName: group?.name || "",
    isOwner: rec.role === "owner",
    isSuperuser: false,
  };
}

export function mapGroup(rec: RecordModel): Group {
  return {
    id: rec.id,
    name: rec.name,
    currency: (rec.currency || "usd").toUpperCase(),
    ...recordDates(rec),
  };
}

export function mapLabel(rec: RecordModel): LabelOut {
  return {
    id: rec.id,
    name: rec.name,
    description: rec.description || "",
    ...recordDates(rec),
  };
}

export function mapLabelSummary(rec: RecordModel): LabelSummary {
  return mapLabel(rec);
}

export function mapProduct(rec: RecordModel): ProductOut {
  return {
    id: rec.id,
    name: rec.name,
    description: rec.description || "",
    manufacturer: rec.manufacturer || "",
    modelNumber: rec.model_number || "",
    ...recordDates(rec),
  };
}

export function mapProductSummary(rec: RecordModel): ProductSummary {
  return {
    id: rec.id,
    name: rec.name,
    manufacturer: rec.manufacturer || "",
    modelNumber: rec.model_number || "",
  };
}

export function mapLocation(rec: RecordModel): LocationOut {
  const parent = rec.expand?.parent;
  return {
    id: rec.id,
    name: rec.name,
    description: rec.description || "",
    parent: parent ? mapLocationSummary(parent) : ({} as LocationSummary),
    children: [],
    ...recordDates(rec),
  };
}

export function mapLocationSummary(rec: RecordModel): LocationSummary {
  return {
    id: rec.id,
    name: rec.name,
    description: rec.description || "",
    ...recordDates(rec),
  };
}

export function mapAttachment(rec: RecordModel): ItemAttachment {
  const file = rec.file as string;
  const fileUrl =
    file && typeof window !== "undefined"
      ? getPb().files.getURL({ id: rec.id, collectionId: COLLECTIONS.attachments } as RecordModel, file)
      : file || "";
  return {
    id: rec.id,
    type: rec.type || "attachment",
    primary: rec.primary || false,
    document: {
      id: rec.id,
      title: rec.title || file || "",
      path: fileUrl,
    },
    ...recordDates(rec),
  };
}

export function mapItemField(rec: RecordModel): ItemField {
  return {
    id: rec.id,
    name: rec.name,
    type: rec.type,
    textValue: rec.text_value || "",
    numberValue: rec.number_value || 0,
    booleanValue: rec.boolean_value || false,
  };
}

export function mapItemSummary(rec: RecordModel): ItemSummary {
  const labels = (rec.expand?.labels as RecordModel[] | undefined) || [];
  const location = rec.expand?.location as RecordModel | undefined;
  const product = rec.expand?.product as RecordModel | undefined;
  const attachments = (rec.expand?.attachments as RecordModel[] | undefined) || [];
  const primaryPhoto = attachments.find(a => a.primary && a.type === "photo") || attachments.find(a => a.type === "photo");

  return {
    id: rec.id,
    name: rec.name || product?.name || "",
    description: rec.description || "",
    quantity: rec.quantity ?? 1,
    insured: rec.insured || false,
    archived: rec.archived || false,
    purchasePrice: String(rec.purchase_price ?? 0),
    labels: labels.map(mapLabelSummary),
    location: location ? mapLocationSummary(location) : null,
    product: product ? mapProductSummary(product) : null,
    imageId: primaryPhoto?.id || "",
    ...recordDates(rec),
  };
}

export function mapItem(rec: RecordModel): ItemOut {
  const summary = mapItemSummary(rec);
  const fields = ((rec.expand?.fields as RecordModel[]) || []).map(mapItemField);
  const attachments = ((rec.expand?.attachments as RecordModel[]) || []).map(mapAttachment);
  const parent = rec.expand?.parent as RecordModel | undefined;

  return {
    ...summary,
    assetId: String(rec.asset_id ?? 0),
    notes: rec.notes || "",
    serialNumber: rec.serial_number || "",
    modelNumber: rec.model_number || "",
    manufacturer: rec.manufacturer || "",
    lifetimeWarranty: rec.lifetime_warranty || false,
    warrantyExpires: toDate(rec.warranty_expires),
    warrantyDetails: rec.warranty_details || "",
    purchaseTime: toDate(rec.purchase_time),
    purchaseFrom: rec.purchase_from || "",
    soldTime: toDate(rec.sold_time),
    soldTo: rec.sold_to || "",
    soldPrice: String(rec.sold_price ?? 0),
    soldNotes: rec.sold_notes || "",
    fields,
    attachments,
    parent: parent ? mapItemSummary(parent) : null,
  };
}

export function mapMaintenance(rec: RecordModel): MaintenanceEntry {
  return {
    id: rec.id,
    name: rec.name,
    description: rec.description || "",
    cost: String(rec.cost ?? 0),
    scheduledDate: toDate(rec.scheduled_date),
    completedDate: toDate(rec.date),
  };
}

export function mapNotifier(rec: RecordModel): NotifierOut {
  return {
    id: rec.id,
    name: rec.name,
    url: rec.url,
    isActive: rec.is_active ?? true,
    ...recordDates(rec),
  };
}

export function itemCreateBody(body: Record<string, any>, groupId: string) {
  return {
    group: groupId,
    name: body.name,
    description: body.description,
    location: body.locationId || "",
    parent: body.parentId || "",
    labels: body.labelIds || [],
    quantity: body.quantity ?? 1,
  };
}

export function itemUpdateBody(body: Record<string, any>) {
  return {
    name: body.name,
    description: body.description,
    location: body.locationId,
    parent: body.parentId,
    labels: body.labelIds,
    quantity: body.quantity,
    insured: body.insured,
    archived: body.archived,
    asset_id: body.assetId ? Number(body.assetId) : undefined,
    serial_number: body.serialNumber,
    model_number: body.modelNumber,
    manufacturer: body.manufacturer,
    lifetime_warranty: body.lifetimeWarranty,
    warranty_expires: body.warrantyExpires,
    warranty_details: body.warrantyDetails,
    purchase_time: body.purchaseTime,
    purchase_from: body.purchaseFrom,
    purchase_price: body.purchasePrice ? Number(body.purchasePrice) : undefined,
    sold_time: body.soldTime,
    sold_to: body.soldTo,
    sold_price: body.soldPrice ? Number(body.soldPrice) : undefined,
    sold_notes: body.soldNotes,
    notes: body.notes,
  };
}
