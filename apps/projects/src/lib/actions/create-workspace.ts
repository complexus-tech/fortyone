"use server";

import type { ApiResponse, Workspace } from "@/types";
import ky from "ky";
import { auth } from "@/auth";
import { getApiUrl } from "@/lib/api-url";
import { buildAuthHeaders } from "@/lib/http/auth-headers";
import { getCookieHeader } from "@/lib/http/header";
import { requestError } from "../fetch-error";

type NewWorkspace = {
  name: string;
  slug: string;
  teamSize: string;
};

const apiUrl = getApiUrl();

export async function createWorkspaceAction(newWorkspace: NewWorkspace) {
  const session = await auth();
  const cookieHeader = await getCookieHeader();
  try {
    const headers = buildAuthHeaders({
      token: session?.token,
      cookieHeader,
    });
    const workspace = await ky
      .post(`${apiUrl}/workspaces`, {
        json: newWorkspace,
        credentials: "include",
        headers,
      })
      .json<ApiResponse<Workspace>>();

    return workspace;
  } catch (error) {
    const data = await requestError<Workspace>(error);
    return data;
  }
}
