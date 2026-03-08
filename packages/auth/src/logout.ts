import { remove } from "api-client";

export const logout = async () => {
  await remove("users/session");
};
