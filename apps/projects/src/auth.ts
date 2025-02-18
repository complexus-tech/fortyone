/* eslint-disable turbo/no-undeclared-env-vars -- ok for now */
/* eslint-disable @typescript-eslint/no-unnecessary-condition -- ok for now */
import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import type { Workspace, UserRole } from "@/types";
import { authenticateUser } from "./lib/actions/users/sigin-in";
import { getWorkspaces } from "./lib/queries/workspaces/get-workspaces";

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
    jwt({ token, user, trigger, session }) {
      if (user) {
        return {
          ...token,
          id: user.id,
          accessToken: user.token,
          lastUsedWorkspaceId: user.lastUsedWorkspaceId,
        };
      }

      if (trigger === "update") {
        token.lastUsedWorkspaceId = session.activeWorkspace.id;
      }

      return token;
    },
    async session({ session, token }) {
      const workspaces = await getWorkspaces(token.accessToken as string);
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
