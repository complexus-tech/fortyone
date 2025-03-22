import type { Metadata } from "next";
import { ListMyStories } from "@/modules/my-work";

export const metadata: Metadata = {
  title: "My Work",
};

export default function Page() {
  return <ListMyStories />;
}
