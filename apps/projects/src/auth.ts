import NextAuth, { CredentialsSignin } from "next-auth";
import Credentials from "next-auth/providers/credentials";
import Google from "next-auth/providers/google";
import {
  authenticateGoogleUser,
  authenticateWithToken,
} from "./lib/actions/auth";
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
      async authorize(credentials, _request) {
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
        return res.data ?? null;
      },
    }),
    Credentials({
      name: "One Tap",
      id: "one-tap",
      credentials: {},
      async authorize(credentials, _request) {
        const { idToken } = credentials as { idToken: string };
        const googleUser = await authenticateGoogleUser({ idToken });
        if (!googleUser) {
          throw new Error("Failed to authenticate Google user");
        }
        return googleUser ?? null;
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
    async jwt({ token, user, account }) {
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
          };
        }
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
    signIn: "/",
    signOut: "/",
    error: `/?error=${encodeURIComponent(errorMessage)}`,
  },
  debug: true,
});
