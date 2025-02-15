"use server";

import { auth, signOut, updateSession } from "@/auth";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";

export const logOut = async (callbackUrl: string) => {
  await signOut({
    redirectTo: `/login?callbackUrl=${encodeURIComponent(callbackUrl)}`,
  });
};

export const changeWorkspace = async (workspaceId: string) => {
  const session = await auth();
  await switchWorkspace(workspaceId);
  const workspace = session?.workspaces.find((w) => w.id === workspaceId);
  await updateSession({
    activeWorkspace: workspace,
    token: session?.token,
  });
};
