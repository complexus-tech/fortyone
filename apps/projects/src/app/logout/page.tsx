import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { signOut } from "@/auth";

export const metadata: Metadata = {
  title: "Login",
};

export default async function Page() {
  await signOut({ redirectTo: "https://www.complexus.app/login" });
  return redirect("https://www.complexus.app/login");
}
