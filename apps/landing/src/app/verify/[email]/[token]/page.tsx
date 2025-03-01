import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { getRedirectUrl } from "@/utils";
import { getMyInvitations } from "@/lib/queries/get-invitations";
import { EmailVerificationCallback } from "./client";

export default async function Page() {
  const session = await auth();
  const invitations = await getMyInvitations();
  if (session) {
    redirect(getRedirectUrl(session, invitations.data || []));
  }
  return <EmailVerificationCallback />;
}
