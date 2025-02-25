import type { Metadata } from "next";
import { CreateAccount } from "@/modules/onboarding/account";

export const metadata: Metadata = {
  title: "Create Account - Complexus",
  description: "Create a new account",
};

export default function CreateAccountPage() {
  return <CreateAccount />;
}
