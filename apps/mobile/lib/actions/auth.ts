import { post } from "@/lib/http";
import { ApiResponse, User } from "@/types";
import { Workspace } from "@/types/workspace";
import ky from "ky";

const apiURL = process.env.EXPO_PUBLIC_API_URL;

type Params = {
  email: string;
  token: string;
};

const getActiveWorkspace = (
  workspaces: Workspace[],
  lastUsedWorkspaceId: string
) => {
  return (
    workspaces.find((workspace) => workspace.id === lastUsedWorkspaceId) ??
    workspaces[0]
  );
};

const getWorkspaces = async () => {
  const response = await ky
    .get(`${apiURL}/workspaces`, {
      credentials: "include",
    })
    .json<ApiResponse<Workspace[]>>();
  return response.data ?? [];
};

export const authenticateWithCode = async (email: string, token: string) => {
  const params = { email, token };
  const response = await post<Params, ApiResponse<User>>(
    "users/verify/email/confirm",
    params,
    {
      useWorkspace: false,
    }
  );
  const workspaces = await getWorkspaces();
  const activeWorkspace = getActiveWorkspace(
    workspaces,
    response.data?.lastUsedWorkspaceId ?? ""
  );

  if (!activeWorkspace) {
    throw new Error("No workspace is available for this account");
  }

  return {
    workspace: activeWorkspace.slug,
  };
};

export const switchWorkspace = async (workspaceId: string) => {
  const response = await post<{ workspaceId: string }, ApiResponse<User>>(
    "workspaces/switch",
    { workspaceId },
    {
      useWorkspace: false,
    }
  );

  return response.data?.lastUsedWorkspaceId;
};
