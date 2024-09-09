"use server";

import { signOut } from "@/auth";

export const logOut = async (callbackUrl: string) => {
  await signOut({
    redirectTo: `/login?callbackUrl=${encodeURIComponent(callbackUrl)}`,
  });
};
