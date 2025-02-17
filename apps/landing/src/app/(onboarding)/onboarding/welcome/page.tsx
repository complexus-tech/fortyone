import type { Metadata } from "next";
import { Welcome } from "@/modules/onboarding/welcome";

export const metadata: Metadata = {
  title: "Welcome - Complexus",
  description: "Welcome to Complexus",
};

export default function WelcomePage() {
  return <Welcome />;
}
