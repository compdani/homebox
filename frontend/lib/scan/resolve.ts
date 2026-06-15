import type { UserClient } from "~~/lib/api/user";

export type ScanEntityType = "product" | "location" | "item" | "asset" | "unknown";

export type ScanEntity = {
  type: ScanEntityType;
  id: string;
  raw: string;
};

function parsePath(pathname: string): ScanEntity | null {
  const parts = pathname.split("/").filter(Boolean);
  if (parts.length < 2) {
    return null;
  }

  const [kind, id] = parts;
  if (kind === "product" || kind === "location" || kind === "item") {
    return { type: kind, id, raw: pathname };
  }
  if (kind === "a" || kind === "assets") {
    return { type: "asset", id, raw: pathname };
  }
  return null;
}

export function resolveScanPayload(raw: string): ScanEntity | null {
  const trimmed = raw.trim();
  if (!trimmed) {
    return null;
  }

  try {
    const url = trimmed.startsWith("http") ? new URL(trimmed) : new URL(trimmed, window.location.origin);
    const fromPath = parsePath(url.pathname);
    if (fromPath) {
      return { ...fromPath, raw: trimmed };
    }
  } catch {
    // fall through
  }

  const pathOnly = trimmed.startsWith("/") ? trimmed : `/${trimmed}`;
  return parsePath(pathOnly);
}

export async function resolveAssetScan(api: UserClient, entity: ScanEntity): Promise<ScanEntity | null> {
  if (entity.type !== "asset") {
    return entity;
  }

  const { data, error } = await api.assets.get(entity.id);
  if (error || !data || data.length !== 1) {
    return null;
  }

  return {
    type: "item",
    id: data[0].id,
    raw: entity.raw,
  };
}
