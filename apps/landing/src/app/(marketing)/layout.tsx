import { cookies } from "next/headers";
import type { ReactNode } from "react";
import { Footer, Navigation } from "@/components/shared";

const SESSION_COOKIE_NAME = "fortyone_session";

export default async function Layout({ children }: { children: ReactNode }) {
  const cookieStore = await cookies();
  const hasSession = cookieStore.has(SESSION_COOKIE_NAME);

  return (
    <>
      <Navigation hasSession={hasSession} />
      {children}
      <Footer />
    </>
  );
}
