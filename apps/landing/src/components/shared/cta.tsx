import { Box, Button } from "ui";
import { Container, UnderlinedHandwrittenAccent } from "@/components/ui";
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
      <Container className="flex items-center justify-center py-24 text-center md:py-36">
        <Box
          className="flex max-w-5xl flex-col items-center"
          data-landing-reveal
        >
          <h2
            className="text-5xl leading-[0.98] font-semibold tracking-tight text-balance md:text-7xl"
            id="marketing-cta-title"
          >
            Ready to make what{" "}
            <UnderlinedHandwrittenAccent tone="primary">
              matters
            </UnderlinedHandwrittenAccent>{" "}
            <UnderlinedHandwrittenAccent tone="success">
              happen?
            </UnderlinedHandwrittenAccent>
          </h2>
          <Box className="mt-8 flex flex-col items-center gap-2 sm:flex-row sm:gap-3">
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
