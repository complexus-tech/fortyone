"use client";

import { useEffect } from "react";

const REVEAL_SELECTOR = "[data-landing-reveal]";

function makeRevealsVisible() {
  document.querySelectorAll<HTMLElement>(REVEAL_SELECTOR).forEach((element) => {
    element.setAttribute("data-visible", "true");
  });
}

export const LandingRevealObserver = () => {
  useEffect(() => {
    const root = document.documentElement;
    const shouldReduceMotion = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    ).matches;

    if (!("IntersectionObserver" in window) || shouldReduceMotion) {
      makeRevealsVisible();
      return;
    }

    const elements = document.querySelectorAll<HTMLElement>(REVEAL_SELECTOR);

    root.classList.add("landing-reveal-ready");

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (!entry.isIntersecting) {
            return;
          }

          entry.target.setAttribute("data-visible", "true");
          observer.unobserve(entry.target);
        });
      },
      {
        rootMargin: "0px 0px -12% 0px",
        threshold: 0.12,
      },
    );

    elements.forEach((element) => {
      observer.observe(element);
    });

    return () => {
      observer.disconnect();
      root.classList.remove("landing-reveal-ready");
    };
  }, []);

  return null;
};
