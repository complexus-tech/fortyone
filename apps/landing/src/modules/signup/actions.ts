"use server";

import { signIn, auth } from "@/auth";

export const logIn = async (formData: FormData) => {
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
};

export const getSession = async () => {
  const session = await auth();
  return session;
};
