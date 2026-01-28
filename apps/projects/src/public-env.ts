import { createPublicEnv } from "next-public-env";

export const { getPublicEnv, PublicEnv } = createPublicEnv(
  {
    NODE_ENV: process.env.NODE_ENV,
    API_URL: process.env.API_URL || process.env.NEXT_PUBLIC_API_URL,
  },
  {
    schema: (z) => ({
      NODE_ENV: z.enum(["development", "production", "test"]),
      API_URL: z.url(),
    }),
  },
);
