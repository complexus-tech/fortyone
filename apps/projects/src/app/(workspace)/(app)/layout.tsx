import type { ReactNode } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { ApplicationLayout } from "@/components/layouts";
import { getQueryClient } from "@/app/get-query-client";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={getQueryClient()}>
      <ApplicationLayout>{children}</ApplicationLayout>
    </QueryClientProvider>
  );
}
