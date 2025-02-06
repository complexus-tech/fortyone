/* eslint-disable @typescript-eslint/no-unnecessary-condition -- ok for now */
import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import ky from "ky";
import { workspaceTags } from "@/constants/keys";
import { DURATION_FROM_SECONDS } from "@/constants/time";
import type { ApiResponse, Workspace, UserRole } from "@/types";
import { authenticateUser } from "./lib/actions/auth/sigin-in";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

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
        revalidate: DURATION_FROM_SECONDS.MINUTE * 20,
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

  callbacks: {
    async jwt({ token, user }) {
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
  pages: {
    signIn: "/login",
    signOut: "/login",
  },
  debug: true,
});
