import React, { useEffect, useRef } from "react";
import { ViewProps, Animated } from "react-native";
import { cn } from "@/lib/utils/classnames";

export interface SkeletonProps extends ViewProps {
  className?: string;
}

export const Skeleton = ({ className, style, ...props }: SkeletonProps) => {
  const opacity = useRef(new Animated.Value(0.3)).current;

  useEffect(() => {
    const animation = Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, {
          toValue: 1,
          duration: 800,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 0.3,
          duration: 800,
          useNativeDriver: true,
        }),
      ])
    );

    animation.start();

    return () => animation.stop();
  }, [opacity]);

  return (
    <Animated.View
      className={cn("bg-gray-100 rounded-lg dark:bg-dark-200", className)}
      style={[{ opacity }, style]}
      {...props}
    />
  );
};
