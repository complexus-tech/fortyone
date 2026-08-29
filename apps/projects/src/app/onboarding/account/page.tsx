import type { Metadata } from "next";
import { CreateAccount } from "@/modules/onboarding/account";

export const metadata: Metadata = {
  title: "Create Account - FortyOne",
  description: "Create a new account",
};

export default async function CreateAccountPage({
  searchParams,
}: {
  searchParams: Promise<{ callbackUrl?: string }>;
}) {
  const { callbackUrl } = await searchParams;

  return <CreateAccount callbackUrl={callbackUrl} />;
}
