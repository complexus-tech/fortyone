import { ApiError, post } from "api-client";

export const logIn = async (email: string, token: string) => {
  try {
    await post("users/verify/email/confirm", {
      email,
      token,
    });

    return { error: null };
  } catch (error) {
    // Verification failures return 400; a 401 here follows a verified email
    // and means this account cannot currently start a session.
    if (error instanceof ApiError && error.status === 401) {
      return { error: "account_unavailable" };
    }
    return {
      error: error instanceof Error ? error.message : "Invalid link",
    };
  }
};
