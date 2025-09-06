import type { Metadata } from "next";
import { AuthLayout } from "@/modules/auth";

export const metadata: Metadata = {
  title: "Signup - FortyOne",
};

export default function Page() {
  return <AuthLayout page="signup" />;
}
