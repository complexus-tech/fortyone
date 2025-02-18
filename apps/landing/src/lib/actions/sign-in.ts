import { signIn } from "@/auth";

export const signInWithGoogle = async (callbackUrl = "/auth-callback") => {
  await signIn("google", {
    redirectTo: callbackUrl,
  });
};
