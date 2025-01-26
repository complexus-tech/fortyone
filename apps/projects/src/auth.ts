import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import { authenticateUser } from "./lib/actions/auth/sigin-in";

type Workspace = {
  id: string;
  name: string;
  isActive: boolean;
};

declare module "next-auth" {
  interface User {
    token: string;
    workspaces: {
      id: string;
      name: string;
      isActive: boolean;
    }[];
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

  callbacks: {
    jwt({ token, user }) {
      if (user) {
        return {
          ...token,
          id: user.id,
          accessToken: user.token,
          workspaces: user.workspaces,
        };
      }
      return token;
    },
    session({ session, token }) {
      const workspaces = token.workspaces as Workspace[];
      const activeWorkspace =
        workspaces.find((w) => w.isActive) || workspaces.at(0);
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
