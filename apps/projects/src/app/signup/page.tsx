import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { SignUpPage } from "@/modules/signup";

export const metadata: Metadata = {
  title: "Sign Up",
};

export default async function Page() {
  const session = await auth();
  if (session) {
    redirect("/");
  }

  return <SignUpPage />;
}
