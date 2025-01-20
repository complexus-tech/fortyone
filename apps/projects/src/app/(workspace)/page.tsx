import { redirect } from "next/navigation";

export default function Page(): JSX.Element {
  return redirect("/my-work");
}
