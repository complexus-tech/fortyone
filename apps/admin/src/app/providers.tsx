"use client";

import type { ReactNode } from "react";
import { useState } from "react";
import { QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from "next-themes";
import { getQueryClient } from "./get-query-client";

export const Providers = ({ children }: { children: ReactNode }) => {
  const [queryClient, setQueryClient] = useState(() => getQueryClient());
  void setQueryClient;

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider attribute="class" enableSystem>
        {children}
      </ThemeProvider>
    </QueryClientProvider>
  );
};
