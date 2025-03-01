/* eslint-disable turbo/no-undeclared-env-vars -- ok for now */

import NextAuth, { CredentialsSignin } from "next-auth";
import Credentials from "next-auth/providers/credentials";
import Google from "next-auth/providers/google";
import type { Workspace, UserRole } from "@/types";
import {
  authenticateGoogleUser,
  authenticateWithToken,
} from "./lib/actions/auth";
import { getWorkspaces } from "./lib/queries/get-workspaces";

const domain =
  process.env.NODE_ENV === "production" ? ".complexus.app" : ".localhost";
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
    Google,
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
        return res.data;
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
    async jwt({ token, user, trigger, session, account, profile }) {
      if (account && user) {
        if (account.provider === "credentials") {
          return {
            ...token,
            id: user.id,
            name: user.name,
            email: user.email,
            picture: user.image,
            accessToken: user.token,
            lastUsedWorkspaceId: user.lastUsedWorkspaceId,
          };
        }
        if (account.provider === "google") {
          const googleUser = await authenticateGoogleUser({
            idToken: account.id_token!,
            email: profile?.email || "",
            fullName: profile?.name || "",
            avatarUrl: profile?.picture || "",
          });
          if (!googleUser) {
            throw new Error("Failed to authenticate Google user");
          }
          return {
            ...token,
            id: googleUser.id,
            picture: googleUser.image,
            name: googleUser.name,
            email: googleUser.email,
            accessToken: googleUser.token,
            lastUsedWorkspaceId: googleUser.lastUsedWorkspaceId,
          };
        }
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
    signOut: "/",
  },
  debug: true,
});
