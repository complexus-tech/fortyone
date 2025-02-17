import type { Metadata } from "next";
import { LoginPage } from "@/modules/login";

export const metadata: Metadata = {
  title: "Login - Complexus",
};

export default function Page() {
  return <LoginPage />;
}
