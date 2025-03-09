"use client";
import { Box, Text, Wrapper } from "ui";
import HeatMap from "@uiw/react-heat-map";

export const Contributions = () => {
  const darkColors = {
    0: "rgb(255 255 255 / 5%)",
    8: "#7BC96F",
    4: "#C6E48B",
    12: "#239A3B",
    32: "#ff7b00",
  };

  function generateContributions(count: number) {
    const contributions = [];
    const daysInMonth = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];

    let currentCount = 0;
    const currentDate = new Date("2023-01-01");

    for (let i = 0; i < 12; i++) {
      const monthDays = daysInMonth[i];
      for (let j = 1; j <= monthDays; j++) {
        const year = currentDate.getFullYear();
        const month = String(currentDate.getMonth() + 1).padStart(2, "0");
        const day = String(currentDate.getDate()).padStart(2, "0");
        const dateString = `${year}/${month}/${day}`;
        const randomCount = Math.floor(Math.random() * 10) + 1; // Random count between 1 and 10
        contributions.push({ date: dateString, count: randomCount });
        currentCount++;
        if (currentCount >= count) {
          return contributions;
        }
        currentDate.setDate(currentDate.getDate() + 1);
      }
    }

    return contributions;
  }
  return (
    <Box>
      <Text fontSize="lg" fontWeight="medium">
        {generateContributions(365).length} contributions in 2023
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
          startDate={new Date("2023/01/01")}
          value={generateContributions(365)}
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
