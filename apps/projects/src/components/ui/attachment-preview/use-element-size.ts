import type { RefObject } from "react";
import { useEffect, useState } from "react";

export const useElementSize = (ref: RefObject<HTMLElement | null>) => {
  const [size, setSize] = useState({ width: 0, height: 0 });
  useEffect(() => {
    const element = ref.current;
    if (!element) return;
    const update = () => {
      const { width, height } = element.getBoundingClientRect();
      setSize((current) =>
        current.width === width && current.height === height
          ? current
          : { width, height },
      );
    };
    update();
    const observer = new ResizeObserver(update);
    observer.observe(element);
    return () => {
      observer.disconnect();
    };
  }, [ref]);
  return size;
};
