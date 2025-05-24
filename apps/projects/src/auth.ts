/* eslint-disable @typescript-eslint/no-unnecessary-condition -- ok for now */

import Credentials from "next-auth/providers/credentials";
import NextAuth, { CredentialsSignin } from "next-auth";
import type { Workspace, UserRole } from "@/types";
import { authenticateWithToken } from "./lib/actions/users/sigin-in";
import { getWorkspaces } from "./lib/queries/workspaces/get-workspaces";
import { DURATION_FROM_SECONDS } from "./constants/time";

const domain =
  process.env.NODE_ENV === "production" ? ".complexus.app" : "localhost";
const useSecureCookies = process.env.NODE_ENV === "production";

class InvalidLoginError extends CredentialsSignin {}
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
        const { email, token } = credentials as {
          email: string;
          token: string;
        };
        const res = await authenticateWithToken({ email, token });
        if (res.error) {
          const error = new InvalidLoginError();
          error.message = res.error.message;
          throw error;
        }
        return res.data!;
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
    jwt({ token, user, trigger, session }) {
      if (user) {
        return {
          ...token,
          id: user.id,
          accessToken: user.token,
          lastUsedWorkspaceId: user.lastUsedWorkspaceId,
          workspaces: user.workspaces,
        };
      }

      if (trigger === "update") {
        if (session.activeWorkspace) {
          token.lastUsedWorkspaceId = session.activeWorkspace.id;
          token.workspaces = session.workspaces;
        } else {
          token.workspaces = session.workspaces;
        }
      }

      return token;
    },
    async session({ session, token }) {
      if (!token.workspaces || (token.workspaces as Workspace[]).length === 0) {
        const workspaces = await getWorkspaces(token.accessToken as string);
        token.workspaces = workspaces;
      }
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
  session: {
    maxAge: DURATION_FROM_SECONDS.MINUTE * 2,
    updateAge: 0,
  },
  jwt: {
    maxAge: DURATION_FROM_SECONDS.MINUTE * 2,
  },
  pages: {
    signIn: "/logout",
    signOut: "/logout",
  },
  debug: process.env.NODE_ENV === "development",
});

export const refreshWorkspaces = async () => {
  await updateSession({ workspaces: undefined });
};
