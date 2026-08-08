import { Box, Button } from "ui";
import { Container } from "@/components/ui";
import { SIGNUP_URL } from "@/lib/app-url";
import { GoogleSignupButton } from "./google-signup-button";

export const CallToAction = () => {
  return (
    <Box
      aria-labelledby="marketing-cta-title"
      as="section"
      className="border-border border-t"
      id="call-to-action"
    >
      <Container className="flex items-center justify-center py-24 text-center md:py-32">
        <Box
          className="flex max-w-2xl flex-col items-center"
          data-landing-reveal
        >
          <h2
            className="text-balances text-4xl leading-[1.02] font-semibold tracking-tight md:text-6xl"
            id="marketing-cta-title"
          >
            Ready to bring strategy, feedback, and work together?
          </h2>
          <Box className="mt-12 flex flex-col items-center gap-2 sm:flex-row sm:gap-3">
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
