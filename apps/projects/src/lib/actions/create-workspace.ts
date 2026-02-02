"use server";

import type { ApiResponse, Workspace } from "@/types";
import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { auth } from "@/auth";
import { requestError } from "../fetch-error";

type NewWorkspace = {
  name: string;
  slug: string;
  teamSize: string;
};

const apiUrl = getApiUrl();

export async function createWorkspaceAction(newWorkspace: NewWorkspace) {
  const session = await auth();
  try {
    const workspace = await ky
      .post(`${apiUrl}/workspaces`, {
        json: newWorkspace,
        headers: {
          Authorization: `Bearer ${session?.token}`,
        },
      })
      .json<ApiResponse<Workspace>>();

    return workspace;
  } catch (error) {
    const data = await requestError<Workspace>(error);
    return data;
  }
}
