import PocketBase from "pocketbase";

let client: PocketBase | null = null;

export function getPb(): PocketBase {
  if (!client) {
    // Root-relative base so API calls stay at /api/... on nested routes (e.g. /locations).
    client = new PocketBase("/");
  }
  return client;
}

/** Re-point the PocketBase client (e.g. integration tests against a running server). */
export function configurePb(baseUrl: string) {
  if (!client) {
    client = new PocketBase(baseUrl);
    return client;
  }
  client.baseUrl = baseUrl;
  return client;
}

export const COLLECTIONS = {
  groups: "hb_groups",
  users: "users",
  locations: "hb_locations",
  labels: "hb_labels",
  products: "hb_products",
  items: "hb_items",
  itemFields: "hb_item_fields",
  attachments: "hb_attachments",
  maintenance: "hb_maintenance_entries",
  notifiers: "hb_notifiers",
  invitations: "hb_group_invitations",
} as const;

export function authGroupId(): string {
  const model = getPb().authStore.model;
  return model?.group || "";
}
