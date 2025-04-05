"use client";
import { Blur, Box, Text } from "ui";
import { KanbanIcon, OKRIcon, SystemIcon, TeamIcon, WorkflowIcon } from "icons";
import { Container } from "@/components/ui";

const Spap = ({
  title,
  overview,
  icon,
}: {
  title: string;
  overview: string;
  icon: React.ReactNode;
}) => (
  <Box>
    {icon}
    <Text
      className="mt-6"
      fontSize="xl"
      fontWeight="medium"
      transform="uppercase"
    >
      {title}
    </Text>

    <Text className="3xl:text-lg mt-4 max-w-xs pl-1 leading-normal opacity-80">
      {overview}
    </Text>
  </Box>
);

const Snapshot = () => {
  const snapshots = [
    {
      title: "Intuitive User Experience",
      icon: <SystemIcon className="h-16" />,
      overview:
        "Beautiful interface that can be learned in a day. No extensive training required.",
    },
    {
      title: "Flexible Workflows",
      icon: <WorkflowIcon className="h-16" />,
      overview:
        "Create custom workflows for each team while maintaining cross-organization visibility.",
    },
    {
      title: "OKR Management",
      icon: <OKRIcon className="h-16" />,
      overview:
        "Set and track Objectives and Key Results (OKRs) to align teams with your vision.",
    },
    {
      title: "Team Collaboration",
      icon: <TeamIcon className="h-16" />,
      overview:
        "Foster teamwork with shared objectives and integrated workspaces that break down silos.",
    },
  ];

  return (
    <Box className="relative mt-16 border-y border-dark-300 bg-black py-12 md:py-24 xl:py-32">
      <Container className="grid-cols-3 gap-12 md:grid">
        <Box>
          <Text
            as="h3"
            className="mb-10 tracking-wide lg:mb-0"
            fontSize="4xl"
            transform="uppercase"
          >
            Why <br /> <span className="text-stroke-white">Choose </span>Us?
          </Text>
          <Text
            className="mt-8 w-11/12 leading-snug opacity-80"
            fontWeight="normal"
          >
            Simplify complexity without sacrifice. Complexus adapts to your
            team&apos;s needs while keeping everyone aligned with strategic
            goals.
          </Text>
        </Box>
        <Box className="col-span-2 grid grid-cols-2 gap-x-8 gap-y-16 lg:gap-x-12 xl:gap-y-36">
          {snapshots.map(({ title, icon, overview }) => (
            <Spap icon={icon} key={title} overview={overview} title={title} />
          ))}
        </Box>
      </Container>
      <div className="pointer-events-none absolute inset-0 z-[3] bg-[url('/noise.png')] bg-repeat opacity-80" />
      <Blur className="absolute bottom-1/2 left-1/2 right-1/2 top-1/2 h-[400px] w-[400px] -translate-x-1/2 -translate-y-1/2 bg-warning/[0.08]" />
    </Box>
  );
};

export const Story = () => {
  return <Snapshot />;
};
