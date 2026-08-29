import { post } from "api-client";

export async function requestMagicEmail(
  email: string,
  isMobileApp: boolean,
  callbackUrl?: string,
) {
  try {
    await post("users/verify/email", {
      email,
      isMobile: isMobileApp,
      callbackURL: callbackUrl,
    });

    return {
      data: null,
      error: {
        message: null,
      },
    };
  } catch (error) {
    return {
      data: null,
      error: {
        message:
          error instanceof Error
            ? error.message
            : "Failed to connect to the server",
      },
    };
  }
}
