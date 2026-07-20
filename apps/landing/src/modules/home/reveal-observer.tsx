"use client";

import { useEffect } from "react";

const REVEAL_SELECTOR = "[data-landing-reveal]";

function makeRevealsVisible() {
  document.querySelectorAll<HTMLElement>(REVEAL_SELECTOR).forEach((element) => {
    element.setAttribute("data-visible", "true");
  });
}

function forEachRevealElement(
  node: Node,
  callback: (element: HTMLElement) => void,
) {
  if (!(node instanceof HTMLElement)) {
    return;
  }

  if (node.matches(REVEAL_SELECTOR)) {
    callback(node);
  }

  node.querySelectorAll<HTMLElement>(REVEAL_SELECTOR).forEach(callback);
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

    const observeElement = (element: HTMLElement) => {
      if (element.dataset.visible !== "true") {
        observer.observe(element);
      }
    };

    elements.forEach(observeElement);

    const mutationObserver = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        mutation.removedNodes.forEach((node) => {
          forEachRevealElement(node, (element) => {
            observer.unobserve(element);
          });
        });

        mutation.addedNodes.forEach((node) => {
          forEachRevealElement(node, observeElement);
        });
      });
    });

    mutationObserver.observe(document.body, {
      childList: true,
      subtree: true,
    });

    return () => {
      mutationObserver.disconnect();
      observer.disconnect();
      root.classList.remove("landing-reveal-ready");
    };
  }, []);

  return null;
};
