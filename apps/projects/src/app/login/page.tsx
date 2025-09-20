import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Login",
};

const domain = process.env.NEXT_PUBLIC_DOMAIN!;
export default async function Page() {
  const session = await auth();
  if (session) {
    redirect("/my-work");
  }

  redirect(`https://${domain}/login`);
}
