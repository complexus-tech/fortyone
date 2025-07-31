import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { ReportCard } from "./components/report-card";

const reports = [
  {
    id: 1,
    title: "Q1 2025 Financial Report",
    description: "Text",
    image: "/images/reports/q1-2025-financial.jpg",
    link: "/reports/q1-2025-financial",
  },
  {
    id: 2,
    title: "Q1 2025 Newsletter",
    description: "Text",
    image: "/images/reports/q1-2025-newsletter.jpg",
    link: "/reports/q1-2025-newsletter",
  },
  {
    id: 3,
    title: "Q4 2024 Financial Report",
    description: "Text",
    image: "/images/reports/q4-2024-financial.jpg",
    link: "/reports/q4-2024-financial",
  },
  {
    id: 4,
    title: "Q3 2025 Financial Report",
    description: "Text",
    image: "/images/reports/q3-2025-financial.jpg",
    link: "/reports/q3-2025-financial",
  },
  {
    id: 5,
    title: "Q3 2025 Newsletter",
    description: "Text",
    image: "/images/reports/q3-2025-newsletter.jpg",
    link: "/reports/q3-2025-newsletter",
  },
  {
    id: 6,
    title: "Q2 2024 Financial Report",
    description: "Text",
    image: "/images/reports/q2-2024-financial.jpg",
    link: "/reports/q2-2024-financial",
  },
];

export const ReportingPage = () => {
  return (
    <Container className="pb-28 pt-12">
      <Text as="h1" className="mb-12 text-5xl font-semibold text-black">
        Reporting
      </Text>
      <Box className="grid grid-cols-1 gap-12 md:grid-cols-3">
        {reports.map((report) => (
          <ReportCard key={report.id} report={report} />
        ))}
      </Box>
    </Container>
  );
};
