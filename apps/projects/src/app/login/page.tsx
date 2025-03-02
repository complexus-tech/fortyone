import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { headers } from "next/headers";
import { auth } from "@/auth";
import { LoginPage } from "@/modules/login";

export const metadata: Metadata = {
  title: "Login",
};

export default async function Page() {
  const session = await auth();
  const headersList = await headers();
  const host = headersList.get("host");

  if (host?.includes("complexus.app")) {
    redirect("https://complexus.app");
  }

  // If user is already logged in, redirect to my work on localhost
  if (session) {
    redirect("/my-work");
  }

  return <LoginPage />;
}
