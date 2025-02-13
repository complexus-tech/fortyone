"use server";

import { revalidatePath } from "next/cache";
import { signOut } from "@/auth";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";

export const logOut = async (callbackUrl: string) => {
  await signOut({
    redirectTo: `/login?callbackUrl=${encodeURIComponent(callbackUrl)}`,
  });
};

export const changeWorkspace = async (workspaceId: string) => {
  await switchWorkspace(workspaceId);

  revalidatePath("/", "layout");
};
