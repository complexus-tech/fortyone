"use server";

import ky from "ky";
import { getApiUrl } from "@/lib/api-url";
import { requestError } from "../fetch-error";

const apiURL = getApiUrl();

export async function requestMagicEmail(email: string, isMobileApp: boolean) {
  try {
    await ky.post(`${apiURL}/users/verify/email`, {
      json: {
        email,
        isMobile: isMobileApp,
      },
    });

    return {
      data: null,
      error: {
        message: null,
      },
    };
  } catch (error) {
    const result = await requestError(error);
    return result;
  }
}
