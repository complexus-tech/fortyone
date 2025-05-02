"use client";
import { Box, Flex, Skeleton, Wrapper } from "ui";

export const ContributionsSkeleton = () => {
  return (
    <Wrapper>
      <Box className="mb-6">
        <Skeleton className="mb-2 h-6 w-40 rounded" />
        <Skeleton className="h-4 w-64 rounded" />
      </Box>

      <Box className="h-[220px] overflow-hidden">
        <Box className="relative h-[220px] w-full">
          {/* Y-axis ticks */}
          <Flex className="absolute left-0 top-0 z-10 h-full flex-col justify-between py-6">
            {Array.from({ length: 4 }).map((_, index) => (
              <Skeleton className="h-3 w-6 rounded" key={index} />
            ))}
          </Flex>
          {/* Chart area */}
          <Box className="absolute inset-0 ml-10 mr-2 mt-4 flex flex-col justify-end">
            {/* Line path simulation */}
            <svg className="w-full overflow-hidden">
              <path
                d="M0,80 C40,50 80,100 120,30 C160,0 200,40 240,20 C280,20 320,60 350,40"
                fill="none"
                stroke="#e2e8f0"
                strokeWidth="2"
              />
              <path
                d="M0,80 C40,50 80,100 120,30 C160,0 200,40 240,20 C280,20 320,60 350,40 L350,140 L0,140 Z"
                fill="url(#skeletonGradient)"
                opacity="0.2"
              />
              <defs>
                <linearGradient
                  id="skeletonGradient"
                  x1="0"
                  x2="0"
                  y1="0"
                  y2="1"
                >
                  <stop offset="0%" stopColor="#e2e8f0" stopOpacity="0.3" />
                  <stop offset="100%" stopColor="#e2e8f0" stopOpacity="0.1" />
                </linearGradient>
              </defs>
            </svg>
            {/* X-axis labels */}
            <Flex className="mt-1 justify-between pr-4">
              {Array.from({ length: 7 }).map((_, index) => (
                <Skeleton className="h-3 w-10 rounded" key={index} />
              ))}
            </Flex>
          </Box>
        </Box>
      </Box>
    </Wrapper>
  );
};
