"use server";
import ky from "ky";
import { auth, updateSession } from "@/auth";
import type { ApiResponse, Workspace } from "@/types";
import { switchWorkspace } from "../users/switch-workspace";

type NewWorkspace = {
  name: string;
  slug: string;
  teamSize: string;
};

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export async function createWorkspaceAction(newWorkspace: NewWorkspace) {
  const session = await auth();
  const workspace = await ky
    .post(`${apiUrl}/workspaces`, {
      json: newWorkspace,
      headers: {
        Authorization: `Bearer ${session?.token}`,
      },
    })
    .json<ApiResponse<Workspace>>();

  await Promise.all([
    switchWorkspace(workspace.data!.id),
    updateSession({
      activeWorkspace: workspace.data,
      token: session?.token,
    }),
  ]);

  return workspace.data;
}
