"use server";
import { AuthError } from "next-auth";
import { signIn } from "@/auth";

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
          return { error: error.message || "Invalid link" };
        default:
          return { error: error.message };
      }
    }
    return {
      error: "Invalid link",
    };
  }
};
