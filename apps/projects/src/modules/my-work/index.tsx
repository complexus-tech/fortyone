"use client";
import { useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { toast } from "sonner";
import type { StoriesLayout } from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";

export const ListMyStories = () => {
  const searchParams = useSearchParams();
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "my-stories:stories:layout",
    "kanban",
  );

  useEffect(() => {
    if (
      searchParams.get("session_id") &&
      !sessionStorage.getItem("stripeSession")
    ) {
      toast.success("Payment successful", {
        description: "Payment for your subscription has been successful",
      });
      sessionStorage.setItem("stripeSession", searchParams.get("session_id")!);
    }
  }, [searchParams]);

  return (
    <MyWorkProvider layout={layout}>
      <Header layout={layout} setLayout={setLayout} />
      <ListMyWork layout={layout} />
    </MyWorkProvider>
  );
};
