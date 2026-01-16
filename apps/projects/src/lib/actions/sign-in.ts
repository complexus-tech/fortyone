"use server";

import { signIn } from "@/auth";

export const signInWithGoogle = async (callbackUrl = "/auth-callback") => {
  await signIn("google", {
    redirectTo: callbackUrl,
  });
};

export const signInWithGoogleOneTap = async (idToken: string) => {
  await signIn("one-tap", {
    redirect: false,
    idToken,
  });
};
