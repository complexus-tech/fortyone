import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const Why = () => {
  return (
    <Container className="pt-16">
      <Box className="mx-auto max-w-6xl">
        <Box className="mb-16 grid grid-cols-1 items-center gap-12 md:grid-cols-5">
          <Text className="text-6xl font-bold md:col-span-2">
            Why join AfricaGiving?
          </Text>
          <Box className="md:col-span-3">
            <Text className="mb-10 text-lg">
              By joining AfricaGiving, your organisation gains access to:
            </Text>
            <ul className="list-disc space-y-4 pl-8">
              <li>
                <Text className="text-lg">
                  <b>A growing network of African givers</b> — individuals,
                  corporates, and diaspora supporters looking to give back.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>A trusted digital giving platform</b> — compliant, secure,
                  and built with African realities in mind.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>Visibility and credibility</b> — be part of a vetted
                  ecosystem of organisations committed to transparency and
                  impact.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>Real-time donation tracking</b> and downloadable reports
                  for easy donor engagement and accountability.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>Marketing support and campaign visibility</b> through our
                  social media, newsletters, and AfricaGiving spotlight
                  features.
                </Text>
              </li>
            </ul>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
