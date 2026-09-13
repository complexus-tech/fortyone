import React, { useEffect } from "react";
import { ViewProps, Animated, useAnimatedValue } from "react-native";
import { useReducedMotion } from "react-native-reanimated";
import { cn } from "@/lib/utils/classnames";

export interface SkeletonProps extends ViewProps {
  className?: string;
}

export const Skeleton = ({ className, style, ...props }: SkeletonProps) => {
  const opacity = useAnimatedValue(0.3);
  const reduceMotion = useReducedMotion();

  useEffect(() => {
    if (reduceMotion) {
      opacity.setValue(0.6);
      return;
    }
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
      ]),
    );

    animation.start();

    return () => animation.stop();
  }, [opacity, reduceMotion]);

  return (
    <Animated.View
      className={cn("bg-gray-100/80 rounded-lg dark:bg-dark-200", className)}
      style={[{ opacity }, style]}
      {...props}
    />
  );
};
