import NextAuth from "next-auth";
import Credentials from "next-auth/providers/credentials";
import { authenticateUser } from "./lib/actions/auth/sigin-in";

export const { handlers, auth, signOut, signIn } = NextAuth({
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
        return {
          ...user,
        };
      },
    }),
  ],
  pages: {
    signIn: "/login",
    signOut: "/login",
  },
  debug: true,
});
