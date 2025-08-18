import type { Metadata } from "next";
import { dehydrate, HydrationBoundary } from "@tanstack/react-query";
import { getQueryClient } from "@/app/get-query-client";
import { getAiChatMessages } from "@/modules/ai-chats/queries/get-ai-chat-messages";
import { aiChatKeys } from "@/modules/ai-chats/constants";
import { auth } from "@/auth";
import { MayaChatPage } from "@/modules/maya/maya";

export const metadata: Metadata = {
  title: "Maya",
  description: "Your AI Assistant",
};

export default async function MayaPage({
  searchParams,
}: {
  searchParams: Promise<{ chatRef: string }>;
}) {
  const session = await auth();
  const { chatRef } = await searchParams;
  const queryClient = getQueryClient();

  if (chatRef && session) {
    await queryClient.prefetchQuery({
      queryKey: aiChatKeys.messages(chatRef),
      queryFn: () => getAiChatMessages(session, chatRef),
    });
  }

  return (
    <HydrationBoundary state={dehydrate(queryClient)}>
      <MayaChatPage />
    </HydrationBoundary>
  );
}
