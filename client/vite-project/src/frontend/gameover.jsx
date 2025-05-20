import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { useState } from "react";
import bugGif from "../assets/image/bug.gif";


function GameOver() {
  const navigate = useNavigate();
  const buttonLabels = ["Restart", "Exit"];

  return (
    <>
        <motion.img 
          src={bugGif} 
          alt="Bug" 
          className="w-max absolute left-1/2 transform -translate-x-1/2 -rotate-180 scale-110 pointer-events-none" 
          initial={{ top:"-100%" }}
          animate={{ top:"120%" }}
          transition={{duration: 3, type: "ease"}}
        />
        
        <div className="h-full flex flex-col items-center justify-center">
            <motion.h1 
              className=" w-full text-6xl sm:text-8xl lg:text-9xl m-5 "
              initial={{ top: "-100%"}}
              animate={{ top:0}}
              transition={{duration: 3, type: "spring", bounce:0.4, delay:1.5 }}
            >
            Game Over
            </motion.h1>
            <div className="w-full flex flex-row gap-10 justify-center">
                {buttonLabels.map((label) => (
                <motion.button
                    className="relative h-full lg:w-1/6 lg:text-xl text-sm" 
                    initial={{ opacity: 0, scale: 0 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{
                      type: "ease",
                      delay: 3,
                      scale: 1,
                    }}
                    whileHover={{scale: 0.9}}
                    onClick={() => navigate("/lobby")}
                >
                {label}
                </motion.button>
                ))}
            </div>
        </div>
    </>
    
  );
};

export default GameOver;
