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
          "At Miningo Tech we aspire to shape the development of geoinformation systems in Africa. We can easily communicate with our clients thanks to Complexus' digital solution, and we're eager to work with them to push the boundaries of what's possible.",
      },
      {
        author: "Shungu Chidovi",
        title: "Founder & Teacher",
        company: "Zimboriginal",
        message: `"Mom, your site is really great, thanks to complexus," my daughter said.  Being able to update my own website, share it with my friends and family, and use it as a teaching tool for my kids and clients makes me really pleased.`,
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
    <Container className="3xl:py-48 relative grid-cols-5 py-20 text-white md:grid xl:border-0 xl:py-36 2xl:py-52">
      <Box className="absolute inset-0 bg-dark">
        <motion.div
          initial={{ width: "20%" }}
          transition={{
            duration: 1,
          }}
          viewport={{ once: true }}
          whileInView={{ width: "100%" }}
        />
      </Box>
      <Box className="absolute inset-0 col-span-2 md:static md:col-span-1 lg:col-span-2">
        <Text
          as="h2"
          className="text-stroke-white introText text-8xl"
          fontWeight="medium"
        >
          Testimonials
        </Text>
      </Box>
      <AnimatePresence>
        <Box className="relative col-span-2 flex flex-col items-start justify-center pl-4 md:col-span-4 lg:col-span-2 xl:pl-0">
          <Text
            as="span"
            className="3xl:mb-16 3xl:text-3xl 3xl:leading-normal relative mb-16 text-lg leading-[2] xl:text-2xl"
          >
            <span className="font-serif 3xl:-left-12 3xl:text-8xl absolute -left-7 -top-6 text-6xl text-primary xl:text-6xl">
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
            className="3xl:mb-4 relative mb-2"
            fontSize="3xl"
            fontWeight="medium"
          >
            {/* eslint-disable-next-line react/jsx-no-comment-textnodes -- ok for this case */}
            <span className="absolute -left-7 top-3 text-primary">//</span>
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
          <Text className="3xl:mb-2 3xl:text-lg mb-1 text-gray-100">
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
          <Text className="3xl:mb-20 mb-12" transform="uppercase">
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
              leftIcon={<ArrowLeft2Icon className="h-6" />}
              onClick={prevSlide}
              rounded="full"
            >
              <span className="sr-only">Previous</span>
            </Button>

            <Button
              asIcon
              className="ml-auto md:h-[3.5rem]"
              color="tertiary"
              leftIcon={<ArrowRight2Icon className="h-6" />}
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
    </Container>
  );
};
