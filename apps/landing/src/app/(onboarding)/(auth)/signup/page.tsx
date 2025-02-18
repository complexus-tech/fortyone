import type { Metadata } from "next";
import { AuthLayout } from "@/modules/auth";

export const metadata: Metadata = {
  title: "Signup - Complexus",
};

export default function Page() {
  return <AuthLayout page="signup" />;
}
