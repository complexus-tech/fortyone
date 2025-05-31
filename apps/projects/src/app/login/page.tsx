import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { auth } from "@/auth";

export const metadata: Metadata = {
  title: "Login",
};

const isLocalhost = process.env.NODE_ENV === "development";

export default async function Page() {
  const session = await auth();
  if (session) {
    redirect("/my-work");
  }

  redirect(
    isLocalhost
      ? "https://complexus.lc/login"
      : "https://www.complexus.app/login",
  );
}
