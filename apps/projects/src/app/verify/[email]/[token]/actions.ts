"use server";

import { AuthError } from "next-auth";
import { auth, signIn } from "@/auth";

export const logIn = async (email: string, token: string) => {
  try {
    await signIn("credentials", {
      email,
      token,
      redirect: false,
    });
  } catch (error) {
    if (error instanceof AuthError) {
      switch (error.type) {
        case "CredentialsSignin":
          return { error: error?.message || "Invalid link" };
        default:
          return { error: error?.message };
      }
    }
    return {
      error: "Invalid link",
    };
  }
};

export const getSession = async () => {
  const session = await auth();
  return session;
};
