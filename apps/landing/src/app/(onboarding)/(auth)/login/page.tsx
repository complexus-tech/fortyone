import type { Metadata } from "next";
import { AuthLayout } from "@/modules/auth";

export const metadata: Metadata = {
  title: "Login - FortyOne",
};

export default function Page() {
  return <AuthLayout page="login" />;
}
