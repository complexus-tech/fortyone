/* eslint-disable @typescript-eslint/no-unnecessary-condition -- ok for now */

import Credentials from "next-auth/providers/credentials";
import NextAuth, { CredentialsSignin } from "next-auth";
import { authenticateWithToken } from "./lib/actions/users/sigin-in";
import { DURATION_FROM_SECONDS } from "./constants/time";

const domain = `.${process.env.NEXT_PUBLIC_DOMAIN!}`;

class InvalidLoginError extends CredentialsSignin {}
declare module "next-auth" {
  interface User {
    token: string;
    lastUsedWorkspaceId: string;
  }
  interface Session {
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
    jwt({ token, user }) {
      if (user) {
        return {
          ...token,
          id: user.id,
          accessToken: user.token,
          lastUsedWorkspaceId: user.lastUsedWorkspaceId,
        };
      }

      return token;
    },
    session({ session, token }) {
      return {
        ...session,
        token: token.accessToken as string,
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
  trustHost: true,
  session: {
    maxAge: DURATION_FROM_SECONDS.DAY * 30,
    updateAge: 0,
  },
  jwt: {
    maxAge: DURATION_FROM_SECONDS.DAY * 30,
  },
  pages: {
    signIn: "/logout",
    signOut: "/logout",
  },
  debug: process.env.NODE_ENV === "development",
});
