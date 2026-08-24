import type { ReactNode } from "react";
import { Footer, Navigation } from "@/components/shared";
import { LandingRevealObserver } from "@/modules/home/reveal-observer";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <>
      <a
        className="bg-background text-foreground focus-visible:ring-ring fixed top-3 left-3 z-50 -translate-y-20 rounded-lg px-4 py-2 shadow-lg transition-transform focus-visible:translate-y-0 focus-visible:ring-2 focus-visible:outline-none motion-reduce:transition-none"
        href="#marketing-content"
      >
        Skip to content
      </a>
      <LandingRevealObserver />
      <Navigation />
      <div className="outline-none" id="marketing-content" tabIndex={-1}>
        {children}
      </div>
      <Footer />
    </>
  );
}
