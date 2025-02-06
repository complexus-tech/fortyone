"use server";
import { redirect } from "next/navigation";
import { signIn } from "@/auth";

export const logIn = async (callbackUrl: string, formData: FormData) => {
  try {
    await signIn("credentials", {
      email: formData.get("email") as string,
      password: formData.get("password") as string,
      redirect: false,
    });
  } catch (error) {
    return {
      error: "Invalid email or password",
    };
  }

  redirect(callbackUrl);
};
