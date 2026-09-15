import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { MOBILE_ACCOUNT_DELETION_PATH } from "@/lib/mobile-auth";
import { DeleteAccountSettings } from "@/modules/settings/account/delete";
import { getLoginUrl } from "@/utils/callback-url";

export const metadata: Metadata = {
  title: "Delete account - FortyOne",
  robots: { index: false, follow: false },
};

export default async function AccountDeletionPage() {
  const session = await auth();
  if (!session) redirect(getLoginUrl(MOBILE_ACCOUNT_DELETION_PATH, true));

  return (
    <main className="mx-auto w-full max-w-3xl px-5 py-10 md:px-8">
      <DeleteAccountSettings
        account={{ id: session.user.id, email: session.user.email }}
        key={session.user.id}
        mobileAuth
      />
    </main>
  );
}
