import type { Metadata } from "next";
import { SignupPage } from "@/modules/signup";

export const metadata: Metadata = {
  title: "Signup - Complexus",
};

export default function Page() {
  return <SignupPage />;
}
