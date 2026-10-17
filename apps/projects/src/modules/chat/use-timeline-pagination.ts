import { useEffect, useRef, useState } from "react";

type TimelinePaginationOptions = {
  enabled: boolean;
  hasOlder: boolean;
  isFetching: boolean;
  error?: Error | null;
  firstItemId?: string;
  positionReady: boolean;
  isNearTop?: () => boolean;
  onLoadOlder?: () => Promise<unknown>;
};

/** Load each history boundary once after the virtual list has positioned itself. */
export const useTimelinePagination = ({
  enabled,
  hasOlder,
  isFetching,
  error,
  firstItemId,
  positionReady,
  isNearTop,
  onLoadOlder,
}: TimelinePaginationOptions) => {
  const [nearTop, setNearTop] = useState(false);
  const [requestPending, setRequestPending] = useState(false);
  const attemptedBoundary = useRef<string | null>(null);
  const mounted = useRef(true);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  useEffect(() => {
    if (
      !enabled ||
      !positionReady ||
      !nearTop ||
      !hasOlder ||
      isFetching ||
      requestPending ||
      error ||
      !firstItemId ||
      !onLoadOlder ||
      attemptedBoundary.current === firstItemId
    )
      return;

    // Prepending restores the scroll anchor after React renders. Recheck the
    // actual viewport next frame rather than trusting its previous top event.
    const frame = requestAnimationFrame(() => {
      if (isNearTop && !isNearTop()) return;
      attemptedBoundary.current = firstItemId;
      setRequestPending(true);
      const finishRequest = () => {
        if (mounted.current) setRequestPending(false);
      };
      // The query owns the visible error and explicit retry. Keep this boundary
      // marked after failure so staying at the top cannot cause a retry loop.
      void Promise.resolve()
        .then(onLoadOlder)
        .then(finishRequest, finishRequest);
    });
    return () => {
      cancelAnimationFrame(frame);
    };
  }, [
    enabled,
    error,
    firstItemId,
    hasOlder,
    isFetching,
    isNearTop,
    nearTop,
    onLoadOlder,
    positionReady,
    requestPending,
  ]);

  return { onNearTopChange: setNearTop };
};
