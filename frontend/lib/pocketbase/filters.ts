import type { ItemsQuery } from "../api/classes/items";

export function buildItemsFilter(q: ItemsQuery = {}): string {
  const parts: string[] = [];

  if (!q.includeArchived) {
    parts.push("archived = false");
  }

  if (q.q) {
    if (q.q.startsWith("#")) {
      const assetId = q.q.slice(1);
      if (assetId) {
        parts.push(`asset_id = ${Number(assetId)}`);
      }
    } else {
      parts.push(`(name ~ "${escapeFilter(q.q)}" || description ~ "${escapeFilter(q.q)}")`);
    }
  }

  if (q.locations?.length) {
    const locs = q.locations.map(id => `location = "${id}"`).join(" || ");
    parts.push(`(${locs})`);
  }

  if (q.labels?.length) {
    for (const id of q.labels) {
      parts.push(`labels ?= "${id}"`);
    }
  }

  if (q.parentIds?.length) {
    const parents = q.parentIds.map(id => `parent = "${id}"`).join(" || ");
    parts.push(`(${parents})`);
  }

  if (q.products?.length) {
    const products = q.products.map(id => `product = "${id}"`).join(" || ");
    parts.push(`(${products})`);
  }

  return parts.join(" && ");
}

function escapeFilter(value: string): string {
  return value.replace(/\\/g, "\\\\").replace(/"/g, '\\"');
}

const SORT_FIELD_MAP: Record<string, string> = {
  createdAt: "created",
  updatedAt: "updated",
};

function mapSortField(field: string): string {
  if (SORT_FIELD_MAP[field]) {
    return SORT_FIELD_MAP[field];
  }
  return field.replace(/([A-Z])/g, "_$1").toLowerCase();
}

export function itemsSort(q: ItemsQuery): string {
  if (!q.orderBy) {
    return "-created";
  }
  const desc = q.orderBy.startsWith("-");
  const field = desc ? q.orderBy.slice(1) : q.orderBy;
  const mapped = mapSortField(field);
  return `${desc ? "-" : ""}${mapped}`;
}
