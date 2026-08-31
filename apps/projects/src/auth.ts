import { SESSION_COOKIE_NAME, getCurrentUser } from "auth";
import { unstable_rethrow as rethrowNextControlFlow } from "next/navigation";
import { cache } from "react";

export type SessionUser = {
  id: string;
  name: string;
  email: string;
  image: string | null;
  username: string;
  fullName: string;
  isInternal: boolean;
  lastUsedWorkspaceId: string;
};

export type Session = {
  user: SessionUser;
  token?: undefined;
};

const resolveSession = async (): Promise<Session | null> => {
  let user: Awaited<ReturnType<typeof getCurrentUser>>;
  try {
    user = await getCurrentUser();
  } catch (error) {
    // Next.js uses thrown control-flow signals to opt routes into dynamic
    // rendering. Preserve those signals even when an upstream client wrapped
    // them as a cause; genuine lookup failures continue to surface unchanged.
    if (typeof window === "undefined") {
      rethrowNextControlFlow(error);
    }
    throw error;
  }

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
      isInternal: user.isInternal,
      lastUsedWorkspaceId: user.lastUsedWorkspaceId,
    },
  };
};

const getServerSession = cache(async (): Promise<Session | null> => {
  // `auth` is also used by browser-side API wrappers. Resolve Next request APIs
  // only on the server, where an AsyncLocalStorage request scope is available.
  const { cookies } = await import("next/headers");
  const sessionCookie = (await cookies()).get(SESSION_COOKIE_NAME)?.value;
  if (!sessionCookie) return null;

  return resolveSession();
});

export const auth = (): Promise<Session | null> =>
  typeof window === "undefined" ? getServerSession() : resolveSession();
