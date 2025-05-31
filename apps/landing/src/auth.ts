import NextAuth, { CredentialsSignin } from "next-auth";
import Credentials from "next-auth/providers/credentials";
import Google from "next-auth/providers/google";
import type { Workspace, UserRole } from "@/types";
import {
  authenticateGoogleUser,
  authenticateWithToken,
} from "./lib/actions/auth";
import { getWorkspaces } from "./lib/queries/get-workspaces";
import { DURATION_FROM_SECONDS } from "./utils";

const domain = `.${process.env.NEXT_PUBLIC_DOMAIN!}`;

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

const errorMessage = "There was an error logging in. Please try again.";

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
      id: "credentials",
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
    Credentials({
      name: "One Tap",
      id: "one-tap",
      credentials: {},
      async authorize(credentials) {
        const { idToken } = credentials as { idToken: string };
        const googleUser = await authenticateGoogleUser({ idToken });
        if (!googleUser) {
          throw new Error("Failed to authenticate Google user");
        }
        return googleUser;
      },
    }),
  ],

  cookies: {
    sessionToken: {
      name: "__Secure-next-auth.session-token",
      options: {
        httpOnly: true,
        sameSite: "lax",
        path: "/",
        domain,
        secure: true,
      },
    },
  },

  callbacks: {
    async jwt({ token, user, trigger, session, account }) {
      if (account && user) {
        if (
          account.provider === "credentials" ||
          account.provider === "one-tap"
        ) {
          return {
            ...token,
            id: user.id,
            name: user.name,
            email: user.email,
            picture: user.image,
            accessToken: user.token,
            lastUsedWorkspaceId: user.lastUsedWorkspaceId,
            workspaces: user.workspaces,
          };
        }
        if (account.provider === "google") {
          const googleUser = await authenticateGoogleUser({
            idToken: account.id_token!,
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
            workspaces: googleUser.workspaces,
          };
        }
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
    maxAge: DURATION_FROM_SECONDS.DAY * 30,
    updateAge: 0,
  },
  jwt: {
    maxAge: DURATION_FROM_SECONDS.DAY * 30,
  },
  pages: {
    signIn: "/login",
    signOut: "/",
    error: `/login?error=${encodeURIComponent(errorMessage)}`,
  },
  debug: true,
});

export const refreshSession = async () => {
  await updateSession({ workspaces: undefined });
};
