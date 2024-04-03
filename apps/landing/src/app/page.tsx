"use client";
import { Badge, Box, Button, Container, Flex, NavLink, Text } from "ui";
import { ArrowRightIcon } from "icons";
import { cn } from "lib";
import "swiper/css";
import "swiper/css/effect-cards";
import { EffectCards } from "swiper/modules";
import { Swiper, SwiperSlide } from "swiper/react";
import Image from "next/image";
import { Logo } from "../components/logo";

const BlurryCircle = ({ className }: { className?: string }) => {
  return (
    <div
      className={cn("h-[300px] w-[300px] rounded-full blur-2xl", className)}
    />
  );
};

export default function Page(): JSX.Element {
  return (
    <>
      <Container className="relative flex h-20 max-w-[85rem] items-center justify-between">
        <Logo className="h-8 dark:text-gray-100" />
        <Flex align="center" gap={6}>
          <NavLink href="/">Features</NavLink>
          <NavLink href="/">Pricing</NavLink>
          <NavLink href="/">Company</NavLink>
          <NavLink href="/">Contact</NavLink>
          <Button className="px-3 text-sm" rounded="full" size="sm">
            Get started
          </Button>
        </Flex>
        <BlurryCircle className="absolute -top-[75vh] left-1/2 right-1/2 h-screen w-[105vw] -translate-x-1/2 bg-primary/15 dark:bg-primary/5" />
      </Container>
      <Container className="max-w-7xl">
        <Flex
          align="center"
          className="md:mt-18 mb-8 mt-16 text-center"
          direction="column"
        >
          <Badge color="tertiary" rounded="full" size="lg">
            Announcing Early Adopters Plan
            <ArrowRightIcon className="h-3 w-auto" />
          </Badge>

          <Text
            as="h1"
            className="mt-6 h-max max-w-6xl pb-2 text-7xl"
            color="gradient"
            fontWeight="medium"
          >
            Empowering teams to conquer project complexity.
          </Text>
          <Text className="mt-4 max-w-[600px] md:mt-6" color="muted">
            Revolutionize project management. Simplify workflows, enhance
            collaboration, achieve exceptional results.
          </Text>
          <Flex align="center" className="mt-8" gap={4}>
            <Button
              className="border border-primary"
              rounded="full"
              size="lg"
              variant="outline"
            >
              Talk to us
            </Button>
            <Button rounded="full" size="lg">
              Get Early Access
            </Button>
          </Flex>
          <Text className="mt-6" color="muted" fontSize="xs">
            No credit card required.
          </Text>
        </Flex>

        <Box className="relative">
          <BlurryCircle className="absolute -top-5 left-1/2 right-1/2 -translate-x-1/2 bg-primary/40 dark:bg-primary/30" />
          <BlurryCircle className="absolute -bottom-20 -left-12 h-[400px] w-[400px] bg-warning/50 dark:bg-warning/10" />
          <BlurryCircle className="absolute -bottom-6 right-0 -z-10 h-[400px] w-[400px] bg-warning/50 dark:bg-warning/20" />
          <Swiper
            effect="cards"
            grabCursor
            initialSlide={1}
            modules={[EffectCards]}
          >
            <SwiperSlide>
              <Image
                alt="Stories"
                className="animate-gradient rounded-2xl bg-gradient-to-br from-primary via-secondary to-warning/60 p-0.5"
                height={2982}
                src="/stories.png"
                width={1854}
              />
            </SwiperSlide>
            <SwiperSlide>
              <Image
                alt="Dashboard"
                className="animate-gradient rounded-2xl bg-gradient-to-br from-primary via-secondary to-warning/60 p-0.5"
                height={2980}
                src="/dashboard.png"
                width={1846}
              />
            </SwiperSlide>
            <SwiperSlide>
              <Image
                alt="Stories"
                className="animate-gradient rounded-2xl bg-gradient-to-br from-primary via-secondary to-warning/60 p-0.5"
                height={2982}
                src="/stories.png"
                width={1854}
              />
            </SwiperSlide>
          </Swiper>
        </Box>
      </Container>
    </>
  );
}
