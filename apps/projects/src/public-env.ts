import { createPublicEnv } from "next-public-env";

export const { getPublicEnv, PublicEnv } = createPublicEnv(
  {
    NODE_ENV: process.env.NODE_ENV,
    API_URL: process.env.NEXT_PUBLIC_API_URL,
    ADMIN_URL: process.env.NEXT_PUBLIC_ADMIN_URL ?? "http://localhost:3002",
  },
  {
    schema: (z) => ({
      NODE_ENV: z.enum(["development", "production", "test"]),
      API_URL: z.url(),
      ADMIN_URL: z.url(),
    }),
  },
);
