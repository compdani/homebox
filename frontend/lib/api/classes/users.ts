import { Requests } from "~~/lib/requests";
import { getPb } from "~~/lib/pocketbase/client";
import { mapUser } from "~~/lib/pocketbase/mappers";
import { wrap, ok } from "~~/lib/pocketbase/response";
import type { ChangePassword, UserOut } from "../types/data-contracts";
import type { Result } from "../types/non-generated";

export class UserApi {
  constructor(private http: Requests) {}

  public self() {
    return wrap(async () => {
      const model = getPb().authStore.model;
      if (!model) {
        throw new Error("unauthorized");
      }
      const rec = await getPb().collection("users").getOne(model.id, { expand: "group" });
      return { item: mapUser(rec) } as Result<UserOut>;
    });
  }

  public logout() {
    getPb().authStore.clear();
    return Promise.resolve(ok<void>(undefined as void));
  }

  public delete() {
    return wrap(async () => {
      const id = getPb().authStore.model?.id;
      if (!id) {
        throw new Error("unauthorized");
      }
      await getPb().collection("users").delete(id);
      getPb().authStore.clear();
    });
  }

  public changePassword(current: string, newPassword: string) {
    return wrap(async () => {
      const id = getPb().authStore.model?.id;
      if (!id) {
        throw new Error("unauthorized");
      }
      await getPb()
        .collection("users")
        .update(id, {
          oldPassword: current,
          password: newPassword,
          passwordConfirm: newPassword,
        });
    });
  }
}
