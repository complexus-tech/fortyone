"use client";
import Image from "next/image";
import { ReactEventHandler, useState } from "react";

import { Box } from "../Box/Box";
import { cn } from "lib";

export const BlurImage = ({
  className = "",
  imageClassName = "",
  priority = false,
  quality = 75,
  alt = "",
  src = "",
  onError,
}: {
  onError?: ReactEventHandler<HTMLImageElement>;
  className?: string;
  imageClassName?: string;
  priority?: boolean;
  quality?: number;
  alt?: string;
  src: string;
}) => {
  const [isLoading, setLoading] = useState(true);

  return (
    <Box
      className={cn(
        "group relative w-full overflow-hidden bg-gray-100 dark:bg-dark-200",
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
          },
          imageClassName
        )}
        onLoad={() => setLoading(false)}
        onError={onError}
      />
    </Box>
  );
};
