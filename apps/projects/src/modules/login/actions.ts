"use server";
import { signIn } from "@/auth";
import { redirect } from "next/navigation";

export const logIn = async (callbackUrl: string, formData: FormData) => {
  try {
    // sleep for 1 second to simulate a network request
    await new Promise((resolve) => setTimeout(resolve, 1000));
    await signIn("credentials", {
      email: "jo@mk.com",
      password: "password",
      redirect: false,
    });
  } catch (error) {
    return {
      error: "Invalid email or password",
    };
  }

  redirect(callbackUrl);
};
