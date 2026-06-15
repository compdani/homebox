import type { PublicApi } from "~~/lib/api/public";
import type { UserOut } from "~~/lib/api/types/data-contracts";
import type { UserClient } from "~~/lib/api/user";
import { getPb } from "~~/lib/pocketbase/client";
import { mapUser } from "~~/lib/pocketbase/mappers";

export interface IAuthContext {
  get token(): boolean | null;
  get attachmentToken(): string | null;
  user?: UserOut;
  isAuthorized(): boolean;
  invalidateSession(): void;
  logout(api: UserClient): ReturnType<UserClient["user"]["logout"]>;
  login(api: PublicApi, email: string, password: string, stayLoggedIn: boolean): ReturnType<PublicApi["login"]>;
}

class AuthContext implements IAuthContext {
  private static _instance?: AuthContext;

  user?: UserOut;

  get token() {
    return getPb().authStore.isValid;
  }

  get attachmentToken() {
    return getPb().authStore.token;
  }

  static get instance() {
    if (!this._instance) {
      this._instance = new AuthContext();
    }
    return this._instance;
  }

  isAuthorized() {
    return getPb().authStore.isValid;
  }

  invalidateSession() {
    this.user = undefined;
    getPb().authStore.clear();
  }

  async login(api: PublicApi, email: string, password: string, stayLoggedIn: boolean) {
    const r = await api.login(email, password, stayLoggedIn);
    if (!r.error) {
      const model = getPb().authStore.model;
      if (model) {
        this.user = mapUser(model);
      }
    }
    return r;
  }

  async logout(api: UserClient) {
    const r = await api.user.logout();
    this.invalidateSession();
    return r;
  }
}

export function useAuthContext(): IAuthContext {
  return AuthContext.instance;
}
