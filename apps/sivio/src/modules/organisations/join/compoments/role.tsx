import { Box, Text } from "ui";
import { Container } from "@/components/ui";

export const Role = () => {
  return (
    <Container className="pt-16">
      <Box className="mx-auto max-w-6xl">
        <Box className="mb-16 grid grid-cols-1 items-center gap-12 md:grid-cols-5">
          <Box className="md:col-span-2">
            <Text className="mb-10 text-6xl font-bold">
              Your Role as a Partner Organisation
            </Text>
            <Text className="text-lg">
              Together, we can grow a culture of giving that is rooted in Africa
              and led by Africans.
            </Text>
          </Box>
          <Box className="md:col-span-3">
            <Text className="mb-10 text-lg">
              As a partner, your responsibility is to:
            </Text>
            <ul className="list-disc space-y-4 pl-8">
              <li>
                <Text className="text-lg">
                  <b>Keep your organisational profile updated</b>, including
                  your mission, work, and impact stories.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>Maintain transparency and trust</b> by submitting
                  up-to-date documentation, including audited financials and
                  reports.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>Drive engagement</b> by sharing your AfricaGiving page with
                  your networks, supporters, and potential donors.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>Build and maintain donor relationships</b> — thank your
                  supporters, keep them informed, and show them the difference
                  they’re making.
                </Text>
              </li>
              <li>
                <Text className="text-lg">
                  <b>
                    Uphold the values of integrity, inclusion, and
                    accountability
                  </b>
                  that define the AfricaGiving ecosystem.
                </Text>
              </li>
            </ul>
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
