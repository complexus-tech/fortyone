import type { Metadata } from "next";
import { CreateAccount } from "@/modules/onboarding/account";

export const metadata: Metadata = {
  title: "Create Account - FortyOne",
  description: "Create a new account",
};

export default function CreateAccountPage() {
  return <CreateAccount />;
}
