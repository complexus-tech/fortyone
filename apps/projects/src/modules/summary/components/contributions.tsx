"use client";
import { Box, Text, Wrapper } from "ui";
import HeatMap from "@uiw/react-heat-map";
import { useContributions } from "@/lib/hooks/contributions";
import { ContributionsSkeleton } from "./contributions-skeleton";

export const Contributions = () => {
  const { data: contributions = [], isPending } = useContributions();
  if (isPending) {
    return <ContributionsSkeleton />;
  }
  const darkColors = {
    0: "rgb(255 255 255 / 5%)",
    8: "#7BC96F",
    4: "#C6E48B",
    12: "#239A3B",
    32: "#ff7b00",
  };

  const mappedContributions = contributions.map((contribution) => {
    const date = new Date(contribution.date);
    const year = date.getFullYear();
    const month = date.getMonth() + 1;
    const day = date.getDate();
    const dateString = `${year}/${month}/${day}`;
    return {
      date: dateString,
      count: contribution.contributions,
    };
  });

  return (
    <Box>
      <Text fontSize="lg" fontWeight="medium">
        {contributions.length} contributions in 2025
      </Text>
      <Wrapper className="mt-2">
        <HeatMap
          className="hidden w-full dark:!text-white"
          legendCellSize={20}
          panelColors={darkColors}
          rectProps={{
            rx: 3,
          }}
          rectSize={15}
          startDate={new Date("2025/01/01")}
          value={mappedContributions}
          weekLabels={["", "Mon", "", "Wed", "", "Fri", ""]}
        />
        <Text className="ml-2.5 mt-4" color="muted" fontSize="sm">
          Light color represents less contributions and dark color represents
          more contributions.
        </Text>
      </Wrapper>
    </Box>
  );
};
