import { route } from "./base";
import { Requests } from "~~/lib/requests";
import { getPb } from "~~/lib/pocketbase/client";
import { ok, wrap } from "~~/lib/pocketbase/response";
import type { APISummary, TokenResponse, UserRegistration } from "./types/data-contracts";

export type StatusResult = {
  health: boolean;
  versions: string[];
  title: string;
  message: string;
};

export class PublicApi {
  constructor(private http: Requests) {}

  public status() {
    return this.http.get<APISummary>({ url: route("/status") });
  }

  public async login(username: string, password: string, _stayLoggedIn = false) {
    try {
      const auth = await getPb().collection("users").authWithPassword(username, password);
      const token = getPb().authStore.token;
      const expiresAt = new Date(auth.meta?.expiresAt || Date.now() + 7 * 24 * 60 * 60 * 1000);
      return ok<TokenResponse>({
        token,
        attachmentToken: token,
        expiresAt,
      });
    } catch {
      return { status: 401, error: true, data: {} as TokenResponse, response: new Response() };
    }
  }

  public register(body: UserRegistration) {
    return wrap(async () => {
      const resp = await fetch(route("/users/register"), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });
      if (!resp.ok) {
        throw new Error("registration failed");
      }
      const data = await resp.json();
      getPb().authStore.save(data.token, data.record);
      return {
        token: data.token,
        attachmentToken: data.token,
        expiresAt: new Date(data.meta?.expiresAt || Date.now() + 7 * 24 * 60 * 60 * 1000),
      } as TokenResponse;
    });
  }
}
