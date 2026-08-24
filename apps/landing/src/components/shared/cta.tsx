import type { ReactNode } from "react";
import { Box, Button } from "ui";
import { cn } from "lib";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";
import { GoogleSignupButton } from "./google-signup-button";

export const CallToAction = ({
  className,
  contentClassName,
  description,
  title = "Ready to bring strategy, feedback, and work together?",
}: {
  className?: string;
  contentClassName?: string;
  description?: ReactNode;
  title?: ReactNode;
}) => {
  return (
    <Box
      aria-labelledby="marketing-cta-title"
      as="section"
      className={cn("border-border border-t", className)}
      id="call-to-action"
    >
      <Container
        className={cn(
          "flex items-center justify-center py-24 text-center md:py-32",
          contentClassName,
        )}
      >
        <Box
          className="flex max-w-2xl flex-col items-center"
          data-landing-reveal
        >
          <h2
            className="text-4xl leading-[1.02] font-semibold tracking-tight text-balance md:text-6xl"
            id="marketing-cta-title"
          >
            {title}
          </h2>
          {description ? (
            <p className="text-text-muted mt-5 max-w-lg text-base leading-relaxed md:text-lg">
              {description}
            </p>
          ) : null}
          <Box
            className={cn(
              "flex flex-col items-center gap-2 sm:flex-row sm:gap-3",
              description ? "mt-8" : "mt-12",
            )}
          >
            <Button
              className="relative z-1 px-3 md:pr-4 md:pl-5"
              color="invert"
              href={SIGNUP_URL}
              rounded="lg"
              size="lg"
            >
              Get started free
            </Button>
            <GoogleSignupButton />
          </Box>
        </Box>
      </Container>
    </Box>
  );
};
