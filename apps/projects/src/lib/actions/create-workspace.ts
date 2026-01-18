"use server";

import ky from "ky";
import { auth } from "@/auth";
import type { ApiResponse, Workspace } from "@/types";
import { requestError } from "../fetch-error";

type NewWorkspace = {
  name: string;
  slug: string;
  teamSize: string;
};

const apiUrl = process.env.NEXT_PUBLIC_API_URL;

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
