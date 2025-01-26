import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { LoginPage } from "@/modules/login";

export const metadata: Metadata = {
  title: "Login",
};

export default async function Page(props: {
  searchParams: Promise<{ callbackUrl: string }>;
}) {
  const searchParams = await props.searchParams;

  const { callbackUrl = "/" } = searchParams;

  const session = await auth();
  if (session) {
    redirect("/");
  }

  return <LoginPage callbackUrl={callbackUrl} />;
}
