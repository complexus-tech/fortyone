"use server";

import ky from "ky";
import { requestError } from "../fetch-error";

const apiURL = process.env.NEXT_PUBLIC_API_URL;

export async function requestMagicEmail(email: string) {
  try {
    await ky.post(`${apiURL}/users/verify/email`, {
      json: {
        email,
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
