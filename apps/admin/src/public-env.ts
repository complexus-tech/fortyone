import { createPublicEnv } from "next-public-env";

export const { getPublicEnv, PublicEnv } = createPublicEnv(
  {
    NODE_ENV: process.env.NODE_ENV,
    API_URL: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8000",
    PROJECTS_URL:
      process.env.NEXT_PUBLIC_PROJECTS_URL ?? "http://localhost:3000",
  },
  {
    schema: (z) => ({
      NODE_ENV: z.enum(["development", "production", "test"]),
      API_URL: z.url(),
      PROJECTS_URL: z.url(),
    }),
  },
);
