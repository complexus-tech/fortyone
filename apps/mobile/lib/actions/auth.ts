import { post } from "@/lib/http";
import { ApiResponse } from "@/types";
import { Workspace } from "@/types/workspace";
import ky from "ky";

const apiURL = process.env.EXPO_PUBLIC_API_URL;

export interface LoginResponse {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string;
  lastUsedWorkspaceId: string;
  token: string;
}
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

const getWorkspaces = async (token: string) => {
  const response = await ky
    .get(`${apiURL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .json<ApiResponse<Workspace[]>>();
  return response.data ?? [];
};

export const authenticateWithToken = async (email: string, token: string) => {
  const params = { email, token };
  const response = await post<Params, ApiResponse<LoginResponse>>(
    "users/verify/email/confirm",
    params,
    {
      useWorkspace: false,
    }
  );
  const workspaces = await getWorkspaces(token);
  const activeWorkspace = getActiveWorkspace(
    workspaces,
    response.data?.lastUsedWorkspaceId ?? ""
  );
  return {
    token: response.data?.token ?? "",
    workspace: activeWorkspace.id,
  };
};
