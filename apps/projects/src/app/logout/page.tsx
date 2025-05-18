import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { logOut } from "@/components/shared/sidebar/actions";

export const metadata: Metadata = {
  title: "Login",
};

export default async function Page() {
  await logOut();
  return redirect("https://www.complexus.app/login");
}
