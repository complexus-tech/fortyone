import { auth } from "@/auth";
import { LoginPage } from "@/modules/login";
import { redirect } from "next/navigation";

export default async function Page(
  props: {
    searchParams: Promise<{ callbackUrl: string }>;
  }
) {
  const searchParams = await props.searchParams;

  const {
    callbackUrl = "/"
  } = searchParams;

  const session = await auth();
  if (session) {
    redirect("/");
  }
  return <LoginPage callbackUrl={callbackUrl} />;
}
