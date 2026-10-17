import type { Metadata } from "next";
import { ChatPage } from "@/modules/chat";

export const metadata: Metadata = {
  title: "Chat",
  description: "Team conversations and direct messages in FortyOne.",
};

export default function Page() {
  return <ChatPage />;
}
