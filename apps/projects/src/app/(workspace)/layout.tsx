import type { Metadata } from "next";
import type { ReactNode } from "react";
import { ApplicationLayout } from "@/components/layouts";
import { StoreProvider } from "@/context/store";
import { getStates } from "@/lib/queries/states/get-states";

export const metadata: Metadata = {
  title: "Objectives",
  description: "Complexus Objectives",
};

export default async function RootLayout({
  children,
}: {
  children: ReactNode;
}) {
  const states = await getStates();

  return (
    <StoreProvider states={states}>
      <ApplicationLayout>{children}</ApplicationLayout>
    </StoreProvider>
  );
}
