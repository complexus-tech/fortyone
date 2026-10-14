import type { MainDetails } from "./components/main-details";

type MainDetailsComponent = typeof MainDetails;

let mainDetailsPromise: Promise<MainDetailsComponent> | undefined;

export const loadMainDetails = (): Promise<MainDetailsComponent> => {
  mainDetailsPromise ??= import("./components/main-details")
    .then((module) => module.MainDetails)
    .catch((error: unknown) => {
      // A transient preload failure must not poison later attempts to open a ticket.
      mainDetailsPromise = undefined;
      throw error;
    });

  return mainDetailsPromise;
};

export const preloadMainDetails = async (): Promise<void> => {
  if (typeof window === "undefined") return;

  try {
    await loadMainDetails();
  } catch {
    // Intent is speculative. Rendering retries through loadMainDetails, where
    // failures still reach the route's error boundary instead of being hidden.
  }
};
