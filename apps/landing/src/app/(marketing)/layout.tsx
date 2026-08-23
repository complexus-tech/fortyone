import type { ReactNode } from "react";
import { Footer, Navigation } from "@/components/shared";
import { LandingRevealObserver } from "@/modules/home/reveal-observer";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <>
      <LandingRevealObserver />
      <Navigation />
      {children}
      <Footer />
    </>
  );
}
