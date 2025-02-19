import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getRedirectUrl } from "@/utils";
import { EmailVerificationCallback } from "./client";

export default async function Page() {
  const session = await auth();
  if (session) {
    redirect(getRedirectUrl(session));
  }
  return <EmailVerificationCallback />;
}
