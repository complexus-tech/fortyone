"use server";

import { revalidatePath } from "next/cache";
import { auth, signOut, updateSession } from "@/auth";
import { switchWorkspace } from "@/lib/actions/users/switch-workspace";

export const logOut = async (callbackUrl: string) => {
  await signOut({
    redirectTo: `/login?callbackUrl=${encodeURIComponent(callbackUrl)}`,
  });
};

export const changeWorkspace = async (workspaceId: string) => {
  const session = await auth();
  const newWorkspaceId = await switchWorkspace(workspaceId);
  const newWorkspace = session?.workspaces.find(
    (workspace) => workspace.id === newWorkspaceId,
  );
  await updateSession({
    activeWorkspace: newWorkspace,
  });
  revalidatePath("/", "layout");
};
