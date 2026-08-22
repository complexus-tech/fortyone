"use client";
import { useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { useLocalStorage, useMediaQuery } from "@/hooks";
import { Header } from "./components/header";
import { ListMyWork } from "./components/list-my-work";
import { MyWorkProvider } from "./components/provider";
import { normalizeMyWorkLayout } from "./types";

export const ListMyStories = () => {
  const searchParams = useSearchParams();
  const isMobile = useMediaQuery("(max-width: 768px)");
  const [storedLayout, setStoredLayout] = useLocalStorage<string>(
    "my-stories:stories:layout",
    isMobile ? "list" : "kanban",
  );
  const layout = normalizeMyWorkLayout(
    storedLayout,
    isMobile ? "list" : "kanban",
  );

  useEffect(() => {
    if (storedLayout !== layout) {
      setStoredLayout(layout);
    }
  }, [layout, setStoredLayout, storedLayout]);

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
    <MyWorkProvider key={layout} layout={layout}>
      <Header
        layout={layout}
        setLayout={(value) => {
          setStoredLayout(value);
        }}
      />
      <ListMyWork layout={layout} />
    </MyWorkProvider>
  );
};
