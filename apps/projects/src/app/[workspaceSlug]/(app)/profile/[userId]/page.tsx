import type { Metadata } from "next";
import { ListUserStories } from "@/modules/profile";

export const metadata: Metadata = {
  title: "Profile",
};

export default async function Page() {
  return <ListUserStories  />;
}
