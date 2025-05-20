import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { useState } from "react";
import bugGif from "../assets/image/bug.gif";


function Home() {
  const navigate = useNavigate();

  return (
    <>
      <motion.img
        src={bugGif}
        alt="Bug"
        className="w-max absolute top-1/2 left-full transform -translate-y-1/2 -rotate-90 scale-110 pointer-events-none"
        initial={{ left: "100%" }}
        animate={{ left: "-120%" }}
        transition={{ duration: 5, type: "ease" }}
      />
      <div className="w-full h-full flex flex-col justify-center items-center">
        <motion.h1
          layoutId="name"
          className="w-max glitch text-6xl sm:text-8xl lg:text-9xl m-5"
          initial={{ left: "100%" }}
          animate={{ left: "0%" }}
          transition={{ duration: 2, type: "ease", delay: 1 }}
        >
          BUG BRAWL
        </motion.h1>
        <motion.button
          className="relative w-max lg:text-xl text-base transform -translate-0"
          initial={{left: "100%" }}
          animate={{ left: "0%"}}
          transition={{ duration: 2, type: "ease", delay: 1, scale: 1 }}
          whileHover={{ scale: 0.9}}
          onClick={() => navigate("/lobby")}>
          Play Game
        </motion.button>
      </div>
    </>
  );
};

export default Home;
