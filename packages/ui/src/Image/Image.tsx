"use client";
import { cn } from "lib";
import Image from "next/image";
import { ReactEventHandler, useState } from "react";

import { Box } from "../Box/Box";

export const BlurImage = ({
  className = "",
  priority = false,
  quality = 75,
  theme = "light",
  alt = "",
  src = "",
  onError,
}: {
  onError?: ReactEventHandler<HTMLImageElement>;
  className?: string;
  priority?: boolean;
  quality?: number;
  theme?: "light" | "dark";
  alt?: string;
  src: string;
}) => {
  const [isLoading, setLoading] = useState(true);

  return (
    <Box
      className={cn(
        "group relative w-full overflow-hidden",
        {
          "bg-gray-100": theme === "light",
          "bg-dark-200": theme === "dark",
        },
        className
      )}
    >
      <Image
        alt={alt}
        src={src}
        priority={priority}
        quality={quality}
        fill
        className={cn(
          "border-separate object-cover object-center scale-100 blur-0 grayscale-0 transition duration-500 ease-in-out",
          {
            "scale-110 blur-2xl grayscale": isLoading,
          }
        )}
        onLoad={() => setLoading(false)}
        onError={onError}
      />
    </Box>
  );
};
