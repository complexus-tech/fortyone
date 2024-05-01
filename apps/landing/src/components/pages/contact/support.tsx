import { Box, Button, Text } from "ui";
import { Container } from "@/components/ui";
import { Blur } from "@/components/ui";
import { EmailIcon, SupportIcon } from "icons";

export const Support = () => {
  const features = [
    {
      heading: "Support & Enquiries",
      email: "support@complexus.tech",
      icon: <SupportIcon className="h-12 w-auto" />,
      description:
        "Contact our support team for assistance with onboarding, troubleshooting or reporting issues.",
    },
    {
      email: "sales@complexus.tech",
      icon: <EmailIcon className="h-12 w-auto" />,
      heading: "Sales",
      description:
        "Contact our sales team for product demos, pricing, and other sales-related queries.",
    },
  ];

  return (
    <Container className="max-w-4xl">
      <Box className="relative">
        <Box className="mb-16 grid grid-cols-1 gap-10 md:mb-32 md:grid-cols-2 md:gap-x-20 md:gap-y-16">
          {features.map(({ heading, description, icon, email }, idx) => (
            <Box
              key={heading}
              className="border-t border-gray-200/10 pt-8 md:pt-10"
            >
              <Text className="mb-4 text-6xl opacity-30 md:text-8xl">
                {icon}
              </Text>
              <Text fontSize="lg" transform="uppercase">
                {heading}
              </Text>
              <a
                className="my-1 inline-block font-semibold opacity-90 transition hover:text-primary"
                href={`mailto:${email}`}
              >
                {email}
              </a>
              <Text fontWeight="normal" className="my-4 text-lg opacity-80">
                {description}
              </Text>
              <Button href={`mailto:${email}`} rounded="full">
                {idx === 0 ? "Contact support" : "Contact sales"}
              </Button>
            </Box>
          ))}
        </Box>
        <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[700px] w-[700px] -translate-x-1/2 -translate-y-1/2 bg-warning/5" />
      </Box>
    </Container>
  );
};
