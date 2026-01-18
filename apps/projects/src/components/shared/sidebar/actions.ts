"use server";

import { redirect } from "next/navigation";
import { signOut } from "@/auth";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";

export const logOut = async () => {
  await signOut();
  redirect("/?signedOut=true");
};

export const changeWorkspace = async (workspaceId: string) => {
  await switchWorkspace(workspaceId);
};
