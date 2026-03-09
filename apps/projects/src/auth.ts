import { getCurrentUser } from "auth";

export type SessionUser = {
  id: string;
  name: string;
  email: string;
  image: string | null;
  username: string;
  fullName: string;
  lastUsedWorkspaceId: string;
};

export type Session = {
  user: SessionUser;
  token?: undefined;
};

const toSession = async (): Promise<Session | null> => {
  const user = await getCurrentUser();

  if (!user) {
    return null;
  }

  return {
    user: {
      id: user.id,
      name: user.fullName || user.username,
      email: user.email,
      image: user.avatarUrl,
      username: user.username,
      fullName: user.fullName,
      lastUsedWorkspaceId: user.lastUsedWorkspaceId,
    },
  };
};

export const auth = async () => {
  return toSession();
};
