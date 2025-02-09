"use server";
import ky from "ky";
import type { ApiResponse } from "@/types";
import { signIn } from "@/auth";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

type SignUpResponse = {
  id: string;
  username: string;
  email: string;
  fullName: string;
  avatarUrl: string;
  lastUsedWorkspaceId: string;
  token: string;
};

export async function signUp(formData: FormData) {
  try {
    const res = await ky
      .post(`${apiURL}/users/register`, {
        json: {
          email: formData.get("email"),
          password: formData.get("password"),
          fullName: formData.get("fullName"),
        },
      })
      .json<ApiResponse<SignUpResponse>>();

    if (res.error) {
      return {
        error: res.error.message,
      };
    }

    // Automatically sign in the user after successful registration
    await signIn("credentials", {
      email: formData.get("email") as string,
      password: formData.get("password") as string,
      redirectTo: "/onboarding/create",
    });

    return { success: true };
  } catch (error) {
    return {
      error: "Failed to create account. Please try again.",
    };
  }
}
