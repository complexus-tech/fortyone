import { useState, useEffect, useCallback } from "react";

export const useMediaQuery = (query: string) => {
  const [matches, setMatches] = useState(false);

  const checkMediaQuery = useCallback(() => {
    const mediaQuery = window.matchMedia(query);
    setMatches(mediaQuery.matches);
  }, [query]);

  useEffect(() => {
    checkMediaQuery();
    const mediaQuery = window.matchMedia(query);

    const handler = (event: MediaQueryListEvent) => {
      setMatches(event.matches);
    };

    mediaQuery.addEventListener("change", handler);

    return () => {
      mediaQuery.removeEventListener("change", handler);
    };
  }, [query, checkMediaQuery]);

  return matches;
};
