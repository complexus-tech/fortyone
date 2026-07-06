"use client";
import Image from "next/image";
import { ReactEventHandler, useState } from "react";

import { Box } from "./box";
import { cn } from "lib";

export const BlurImage = ({
  className = "",
  imageClassName = "",
  priority = false,
  quality = 75,
  sizes = "100vw",
  unoptimized = false,
  alt = "",
  src = "",
  onError,
}: {
  onError?: ReactEventHandler<HTMLImageElement>;
  className?: string;
  imageClassName?: string;
  priority?: boolean;
  quality?: number;
  sizes?: string;
  unoptimized?: boolean;
  alt?: string;
  src: string;
}) => {
  const [isLoading, setLoading] = useState(true);

  return (
    <Box
      className={cn(
        "group relative w-full overflow-hidden bg-surface-muted",
        className
      )}
    >
      <Image
        alt={alt}
        src={src}
        priority={priority}
        quality={quality}
        sizes={sizes}
        unoptimized={unoptimized}
        fill
        className={cn(
          "border-separate object-cover object-center scale-100 blur-0 grayscale-0 transition duration-500 ease-in-out",
          {
            "scale-110 blur-2xl grayscale": isLoading,
          },
          imageClassName
        )}
        onLoad={() => setLoading(false)}
        onError={onError}
      />
    </Box>
  );
};
