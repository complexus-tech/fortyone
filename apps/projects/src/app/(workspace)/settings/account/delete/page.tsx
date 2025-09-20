import type { Metadata } from "next";
import { DeleteAccountSettings } from "@/modules/settings/account/delete";

export const metadata: Metadata = {
  title: "Settings › Account › Delete",
};

export default function Page() {
  return <DeleteAccountSettings />;
}
