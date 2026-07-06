import ky from "ky";
import { ApiResponse, User } from "@/types";

const apiURL = process.env.EXPO_PUBLIC_API_URL;

export const getCurrentUser = async () => {
  const response = await ky
    .get(`${apiURL}/auth/me`, {
      credentials: "include",
    })
    .json<ApiResponse<User>>();

  return response.data;
};

export const clearSession = async () => {
  await ky.delete(`${apiURL}/users/session`, {
    credentials: "include",
  });
};
