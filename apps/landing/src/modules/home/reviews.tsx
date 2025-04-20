"use client";
import { motion, AnimatePresence } from "framer-motion";
import { Container, Box, Text, Button, Flex } from "ui";
import { useState, useCallback, useMemo, useEffect } from "react";
import "./reviews.css";
import { ArrowLeft2Icon, ArrowRight2Icon } from "icons";

export const Testimonials = () => {
  const slides = useMemo(
    () => [
      {
        author: "Tshaxedue Gondo",
        title: "Founder & Spatial Data Scientist",
        company: "Miningo Technologies",
        message:
          "At Miningo Tech, managing multiple projects across Africa was challenging until we found this solution. The project tracking system has revolutionized how we handle our geoinformation systems development. Team collaboration has never been smoother.",
      },
      {
        author: "Shungu Chidovi",
        title: "Founder & Teacher",
        company: "Zimboriginal",
        message:
          "The project management platform has transformed how I organize my teaching materials and coordinate with other educators. Being able to track progress, set milestones, and share resources with my team has made a tremendous difference in our educational projects.",
      },
    ],
    [],
  );
  const [slide, setSlide] = useState(slides[0]);

  const nextSlide = useCallback(() => {
    const nextIndex = slides.indexOf(slide) + 1;
    if (nextIndex >= slides.length) {
      setSlide(slides[0]);
    } else {
      setSlide(slides[nextIndex]);
    }
  }, [slide, slides]);

  const prevSlide = useCallback(() => {
    const prevIndex = slides.indexOf(slide) - 1;
    if (prevIndex < 0) {
      setSlide(slides[slides.length - 1]);
    } else {
      setSlide(slides[prevIndex]);
    }
  }, [slide, slides]);

  useEffect(() => {
    const interval = setInterval(() => {
      nextSlide();
    }, 8000);
    return () => {
      clearInterval(interval);
    };
  }, [slide, slides, nextSlide]);

  return (
    <Container className="3xl:py-48 relative grid-cols-5 overflow-hidden bg-gradient-to-b from-dark-300 via-black via-40% to-black py-28 text-white md:grid md:h-[95vh] xl:py-36 2xl:py-52">
      <Box className="pointer-events-none absolute inset-0 col-span-2 md:static md:col-span-1 lg:col-span-2">
        <Text
          as="h2"
          className="text-stroke-white introText hidden text-6xl md:block md:text-8xl"
          fontWeight="medium"
        >
          Testimonials
        </Text>
      </Box>
      <AnimatePresence>
        <Box className="relative col-span-2 flex flex-col items-start justify-center pl-4 md:col-span-4 lg:col-span-2 xl:pl-0">
          <Text
            as="span"
            className="relative mb-16 text-xl font-normal md:text-3xl md:leading-snug"
          >
            <span className="3xl:-left-12 3xl:text-8xl absolute -left-7 -top-6 text-6xl text-primary xl:text-6xl">
              &ldquo;
            </span>
            <motion.div
              animate="animateState"
              exit="exitState"
              initial="initialState"
              key={slide.author + slide.company}
              transition={{
                duration: 0.75,
                type: "spring",
                damping: 10,
              }}
              variants={{
                initialState: {
                  y: 15,
                },
                animateState: {
                  y: 0,
                },
                exitState: {
                  y: -15,
                },
              }}
            >
              {slide.message}
            </motion.div>
          </Text>

          <Text
            as="span"
            className="3xl:mb-4 relative mb-2 block text-2xl md:text-3xl"
            fontWeight="normal"
          >
            {/* eslint-disable-next-line react/jsx-no-comment-textnodes -- ok for this case */}
            <span className="absolute -left-7 text-primary">//</span>
            <motion.div
              animate="animateState"
              exit="exitState"
              initial="initialState"
              key={slide.author + slide.company}
              transition={{
                duration: 0.65,
                delay: 0.15,
                type: "spring",
                damping: 10,
              }}
              variants={{
                initialState: {
                  y: 10,
                },
                animateState: {
                  y: 0,
                },
                exitState: {
                  y: -10,
                },
              }}
            >
              {slide.author}
            </motion.div>
          </Text>
          <Text
            as="span"
            className="mb-1 block text-gray-100 md:mb-2 md:text-xl"
          >
            <motion.div
              animate="animateState"
              exit="exitState"
              initial="initialState"
              key={slide.author + slide.company}
              transition={{
                duration: 0.65,
                delay: 0.2,
                type: "spring",
                damping: 10,
              }}
              variants={{
                initialState: {
                  y: 10,
                },
                animateState: {
                  y: 0,
                },
                exitState: {
                  y: -10,
                },
              }}
            >
              {slide.title}
            </motion.div>
          </Text>
          <Text
            as="span"
            className="3xl:mb-20 mb-8 block text-lg md:mb-12"
            color="muted"
            transform="uppercase"
          >
            <motion.div
              animate="animateState"
              exit="exitState"
              initial="initialState"
              key={slide.author + slide.company}
              transition={{
                duration: 0.65,
                delay: 0.22,
                type: "spring",
                damping: 10,
              }}
              variants={{
                initialState: {
                  y: 10,
                },
                animateState: {
                  y: 0,
                },
                exitState: {
                  y: -10,
                },
              }}
            >
              {slide.company}
            </motion.div>
          </Text>
          <Flex className="gap-5">
            <Button
              asIcon
              className="ml-auto md:h-[3.5rem]"
              color="tertiary"
              leftIcon={<ArrowLeft2Icon className="md:h-6" />}
              onClick={prevSlide}
              rounded="full"
            >
              <span className="sr-only">Previous</span>
            </Button>

            <Button
              asIcon
              className="ml-auto md:h-[3.5rem]"
              color="tertiary"
              leftIcon={<ArrowRight2Icon className="md:h-6" />}
              onClick={nextSlide}
              rounded="full"
            >
              <span className="sr-only">Next</span>
            </Button>
          </Flex>

          <Box className="3xl:-bottom-32 absolute -bottom-12 right-6 md:-right-40 xl:-bottom-16 xl:-right-60">
            <svg
              className="3xl:w-40 h-auto w-20 rotate-6 text-white opacity-30 xl:w-32 2xl:w-36"
              fill="none"
              height="470"
              viewBox="0 0 266 470"
              width="266"
              xmlns="http://www.w3.org/2000/svg"
            >
              <line
                stroke="currentColor"
                x1="202.44"
                x2="0.43993"
                y1="33.2376"
                y2="407.238"
              />
              <line
                stroke="currentColor"
                x1="265.439"
                x2="10.4393"
                y1="0.238835"
                y2="469.239"
              />
              <circle cx="123.5" cy="212.5" r="86" stroke="currentColor" />
            </svg>
          </Box>
        </Box>
      </AnimatePresence>
      {/* <div className="pointer-events-none absolute inset-0 z-[3] bg-[url('/noise.png')] bg-repeat opacity-60" /> */}
    </Container>
  );
};
