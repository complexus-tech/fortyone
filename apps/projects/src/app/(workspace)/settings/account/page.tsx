import type { Metadata } from "next";
import { ProfileSettings } from "@/modules/settings/account/profile";

export const metadata: Metadata = {
  title: "Settings › Profile",
};

export default function Page() {
  return <ProfileSettings />;
}
