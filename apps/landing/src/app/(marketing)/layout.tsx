import type { ReactNode } from "react";
import { CallToAction, Footer, Navigation } from "@/components/shared";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <>
      <Navigation />
      {children}
      <CallToAction />
      <Footer />
    </>
  );
}
