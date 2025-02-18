import { redirect } from "next/navigation";
import { auth } from "@/auth";
import { ClientPage } from "./client";

export default async function AuthCallback() {
  const session = await auth();
  if (!session) {
    redirect("/login");
  }
  return <ClientPage />;
}
