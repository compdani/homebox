import { getPb } from "~~/lib/pocketbase/client";

export async function downloadAuthedFile(url: string, filename: string): Promise<boolean> {
  const token = getPb().authStore.token;
  const headers: Record<string, string> = {};
  if (token) {
    headers.Authorization = token;
  }

  const response = await fetch(url, { headers });
  if (!response.ok) {
    return false;
  }

  const blob = await response.blob();
  const objectUrl = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = objectUrl;
  link.download = filename;
  link.click();
  URL.revokeObjectURL(objectUrl);
  return true;
}
