import { post } from "api-client";

export const logIn = async (email: string, token: string) => {
  try {
    await post("users/verify/email/confirm", {
      email,
      token,
    });

    return { error: null };
  } catch (error) {
    return {
      error: error instanceof Error ? error.message : "Invalid link",
    };
  }
};
