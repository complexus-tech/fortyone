import { useState } from "react";
import { getBoundAccountId, native } from "../native";

export function useAccountNative() {
  return useState(() => {
    const accountId = getBoundAccountId();
    return { ...native.forAccount(accountId), accountId };
  })[0];
}
