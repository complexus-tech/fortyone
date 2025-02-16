"use server";
import ky from "ky";
import { revalidateTag } from "next/cache";
import { auth } from "@/auth";
import type { ApiResponse, Workspace } from "@/types";
import { workspaceTags } from "@/constants/keys";
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
  revalidateTag(workspaceTags.lists());

  await switchWorkspace(workspace.data!.id);

  return workspace.data;
}
