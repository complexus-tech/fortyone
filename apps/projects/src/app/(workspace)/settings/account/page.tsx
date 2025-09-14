import type { Metadata } from "next";
import { ProfileSettings } from "@/modules/settings/account/profile";

export const metadata: Metadata = {
  title: "Settings › Account",
};

export default function Page() {
  return <ProfileSettings />;
}
