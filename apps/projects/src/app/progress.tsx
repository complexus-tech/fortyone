"use client";

import { useEffect } from "react";
import NProgress from "nprogress";

type PushStateInput = [
  data: unknown,
  unused: string,
  url?: string | URL | null | undefined,
];

export const ProgressBar = () => {
  useEffect(() => {
    NProgress.configure({ showSpinner: false });

    const handleAnchorClick = (event: MouseEvent) => {
      const targetUrl = (event.currentTarget as HTMLAnchorElement).href;
      const currentUrl = window.location.href;
      if (targetUrl !== currentUrl) {
        NProgress.start();
      }
    };

    const handleMutation: MutationCallback = () => {
      const anchorElements: NodeListOf<HTMLAnchorElement> =
        document.querySelectorAll("a[href]");

      anchorElements.forEach((anchor) => {
        anchor.addEventListener("click", handleAnchorClick);
      });
    };

    const mutationObserver = new MutationObserver(handleMutation);

    mutationObserver.observe(document, { childList: true, subtree: true });

    // eslint-disable-next-line @typescript-eslint/unbound-method -- Proxy is a constructor
    window.history.pushState = new Proxy(window.history.pushState, {
      apply: (target, thisArg, argArray: PushStateInput) => {
        NProgress.done();
        target.apply(thisArg, argArray);
      },
    });
  });

  return <div />;
};
