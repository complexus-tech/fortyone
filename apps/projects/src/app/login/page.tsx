import { auth } from "@/auth";
import { LoginPage } from "@/modules/login";
import { redirect } from "next/navigation";

export default async function Page({
  searchParams: { callbackUrl = "/" },
}: {
  searchParams: { callbackUrl: string };
}) {
  const session = await auth();
  if (session) {
    redirect("/");
  }
  return <LoginPage callbackUrl={callbackUrl} />;
}
