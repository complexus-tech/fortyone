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
          "flex items-center justify-center py-16 text-center sm:py-24 md:py-32",
          contentClassName,
        )}
      >
        <Box
          className="flex max-w-2xl flex-col items-center"
          data-landing-reveal
        >
          <h2
            className="text-3xl leading-[1.02] font-semibold tracking-tight text-balance sm:text-4xl md:text-6xl"
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
              "flex w-full max-w-[18rem] flex-col items-stretch gap-2 sm:w-auto sm:max-w-none sm:flex-row sm:items-center sm:gap-3 [&>a]:w-full [&>a]:justify-center sm:[&>a]:w-max",
              description ? "mt-8" : "mt-12",
            )}
          >
            <Button
              className="relative z-1 w-full justify-center px-3 sm:w-max md:pr-4 md:pl-5"
              color="invert"
              href={SIGNUP_URL}
              rounded="md"
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
