import type { RecordModel } from "pocketbase";
import type { TResponse } from "~~/lib/requests";

export function ok<T>(data: T, status = 200): TResponse<T> {
  return { status, error: false, data, response: new Response() };
}

export function err<T>(status = 400): TResponse<T> {
  return { status, error: true, data: {} as T, response: new Response() };
}

export async function wrap<T>(fn: () => Promise<T>): Promise<TResponse<T>> {
  try {
    const data = await fn();
    return ok(data);
  } catch (e: any) {
    const status = e?.status || 400;
    return err(status);
  }
}

export function toDate(value?: string): Date | string {
  if (!value) {
    return "";
  }
  return new Date(value);
}

export function recordDates(rec: RecordModel) {
  return {
    createdAt: toDate(rec.created),
    updatedAt: toDate(rec.updated),
  };
}
