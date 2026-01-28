import { getPublicEnv } from "@/public-env";

export const getApiUrl = () => {
  if (typeof window === "undefined") {
    return process.env.INTERNAL_API_URL ?? process.env.API_URL ?? "";
  }

  return getPublicEnv().API_URL;
};
