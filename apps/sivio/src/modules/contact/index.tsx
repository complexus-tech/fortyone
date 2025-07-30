import { Box } from "ui";
import { Container } from "@/components/ui";
import { VoiceSection } from "./components/voice-section";
import { ContactForm } from "./components/contact-form";
import { NewsletterForm } from "./components/newsletter-form";

export const ContactPage = () => {
  return (
    <Box>
      <Container className="py-2">
        <VoiceSection />
        <Box className="grid grid-cols-1 gap-8 md:grid-cols-2">
          <ContactForm />
          <NewsletterForm />
        </Box>
      </Container>
    </Box>
  );
};
