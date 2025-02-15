/* eslint-disable turbo/no-undeclared-env-vars -- ok for now */
/* eslint-disable @typescript-eslint/no-unnecessary-condition -- ok for now */
import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import ky from "ky";
import type { ApiResponse, Workspace, UserRole } from "@/types";
import { authenticateUser } from "./lib/actions/users/sigin-in";
import { DURATION_FROM_SECONDS } from "./constants/time";
import { workspaceTags } from "./constants/keys";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

const domain =
  process.env.NODE_ENV === "production" ? ".complexus.app" : ".localhost";
const useSecureCookies = process.env.NODE_ENV === "production";
declare module "next-auth" {
  interface User {
    token: string;
    lastUsedWorkspaceId: string;
    workspaces: Workspace[];
    userRole: UserRole;
  }
  interface Session {
    workspaces: Workspace[];
    activeWorkspace: Workspace;
    token: string;
  }
}

const getWorkspaces = async (token: string) => {
  const workspaces = await ky
    .get(`${apiURL}/workspaces`, {
      headers: { Authorization: `Bearer ${token}` },
      next: {
        revalidate: DURATION_FROM_SECONDS.MINUTE * 10,
        tags: [workspaceTags.lists()],
      },
    })
    .json<ApiResponse<Workspace[]>>();
  return workspaces.data!;
};

export const {
  handlers,
  auth,
  signOut,
  signIn,
  unstable_update: updateSession,
} = NextAuth({
  providers: [
    Credentials({
      name: "Credentials",
      credentials: {},
      async authorize(credentials) {
        const { email, password } = credentials as {
          email: string;
          password: string;
        };
        const user = await authenticateUser({ email, password });
        if (!user) {
          throw new Error("Invalid email or password");
        }
        return user;
      },
    }),
  ],

  cookies: {
    sessionToken: {
      name: `${useSecureCookies ? "__Secure-" : ""}next-auth.session-token`,
      options: {
        httpOnly: true,
        sameSite: "lax",
        path: "/",
        domain: useSecureCookies ? domain : undefined,
        secure: useSecureCookies,
      },
    },
  },

  callbacks: {
    async jwt({ token, user, trigger, session }) {
      if (user) {
        const workspaces = await getWorkspaces(user.token);
        return {
          ...token,
          id: user.id,
          accessToken: user.token,
          lastUsedWorkspaceId: user.lastUsedWorkspaceId,
          workspaces,
        };
      }

      if (trigger === "update") {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-argument -- ok for now
        const workspaces = await getWorkspaces(session.token);
        token.lastUsedWorkspaceId = session.activeWorkspace.id;
        token.workspaces = workspaces;
      }

      return token;
    },
    session({ session, token }) {
      const workspaces = token.workspaces as Workspace[];
      const activeWorkspace =
        workspaces.find((w) => w.id === token.lastUsedWorkspaceId) ||
        workspaces.at(0);
      return {
        ...session,
        token: token.accessToken,
        workspaces,
        activeWorkspace,
        user: {
          ...session.user,
          id: token.id as string,
          name: token.name,
          email: token.email,
          image: token.picture,
          userRole: activeWorkspace?.userRole || "guest",
        },
      };
    },
  },
  trustHost: true,
  pages: {
    signIn: "/login",
    signOut: "/login",
  },
  debug: true,
});
