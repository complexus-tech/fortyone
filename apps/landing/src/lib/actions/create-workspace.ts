"use server";
import ky from "ky";
import { auth } from "@/auth";
import type { ApiResponse, Workspace, User } from "@/types";

type NewWorkspace = {
  name: string;
  slug: string;
  teamSize: string;
};

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

export async function switchWorkspace(workspaceId: string) {
  const session = await auth();
  const res = await ky.post(`${apiUrl}/workspaces/switch`, {
    json: {
      workspaceId,
    },
    headers: {
      Authorization: `Bearer ${session?.token}`,
    },
  });

  const user = await res.json<ApiResponse<User>>();

  return user.data;
}

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

  await switchWorkspace(workspace.data!.id);

  return workspace.data;
}
