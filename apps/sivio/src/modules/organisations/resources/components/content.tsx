import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { ResourceCard } from "./resource-card";

const resources = [
  {
    id: 1,
    title: "School for Sustainable CSOs",
    subtitle: "Online Training Course",
    image: "/images/about/4.jpg",
  },
  {
    id: 2,
    title: "Building Sustainable NGOs",
    subtitle:
      "Udemy Masterclass: Harnessing individual giving for sustainable change",
    image: "/images/about/4.jpg",
  },
  {
    id: 3,
    title: "Reporting Templates",
    image: "/images/about/4.jpg",
  },
  {
    id: 4,
    title: "Quarterly Reports",
    smallText: "2020 / 2021",
    image: "/images/about/4.jpg",
  },
];

export const Content = () => {
  return (
    <Container className="pt-16">
      <Box className="mx-auto max-w-6xl">
        <Box className="mb-16 grid grid-cols-1 gap-12 md:grid-cols-5">
          <Text className="text-5xl font-bold md:col-span-2">
            Resources to strengthen your fundraising efforts
          </Text>
          <Text className="text-lg md:col-span-3">
            AfricaGiving offers a growing library of tools to support your
            fundraising strategy. Access our Udemy Masterclass{" "}
            <strong>
              Building Sustainable NGOs - harnessing individual giving for
              sustainable change
            </strong>
            , enrol in our online course on building sustainable CSOs, and
            download practical templates and reporting guides—all designed to
            support impactful, transparent, and strategic giving.
          </Text>
        </Box>

        <Box className="grid grid-cols-1 md:grid-cols-2">
          {resources.map((resource) => (
            <ResourceCard key={resource.id} {...resource} />
          ))}
        </Box>
      </Box>
    </Container>
  );
};
