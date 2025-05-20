import React, { useState } from 'react';
import { useNavigate } from "react-router-dom";
import { delay, motion } from "framer-motion";
import Popup from "../widget/popup";

import Sabotage from '../widget/sabotage.jsx';

export default function Lobby() {
  const buttonLabels = ["Create Match", "Join Match", "Find Match"];
  const [selectedLabel, setSelectedLabel] = useState('');
  const [showPopup, setShowPopup] = useState(false);
  const navigate = useNavigate();

  const handleButtonClick = (label) => {
    setSelectedLabel(label);
    setShowPopup(true);

    if (label === "Find Match") {
      setTimeout(() => {
        setShowPopup(false);
        navigate("/game");
      }, 2000);
    }
  };

  const handleClosePopup = () => {
    setShowPopup(false);
    setSelectedLabel('');
  };

  //Show all efect
  const [showEffectScreen, setShowEffectScreen] = useState(false);
  const handleTestEffectClick = () => {
    setShowEffectScreen(true);
    setTimeout(() => setShowEffectScreen(false), 60000);
  };

  //Flicker effect
  const [showFlicker, setShowFlicker] = useState(false);
  const handleFlickerClick = () => {
    setShowFlicker(true);
    setTimeout(() => setShowFlicker(false), 10000);
  };

  //Blurry effect
  const [showBlurry, setShowBlurry] = useState(false);
  const handleBlurryClick = () => {
    setShowBlurry(true);
    setTimeout(() => setShowBlurry(false), 10000);
  };

  //Glitch effect
  const [showGlitch, setShowGlitch] = useState(false);
  const handleGlitchClick = () => {
    setShowGlitch(true);
    setTimeout(() => setShowBlurry(false), 10000);
  };

  const [isBackwards, setIsBackwards] = useState(false);
  const handleBackwardsClick = () => {
    setIsBackwards(true);
    setTimeout(() => setIsBackwards(false), 10000);
  };

  return (
    <div className="h-full w-full flex flex-col gap-3 justify-center items-center">
      <motion.h1
        layoutId="name"
        className="lg:text-9xl text-8xl"
        initial={{ scale: 1 }}
        animate={{ scale: 0.8 }}
        transition={{ type: "spring", duration: 1 }}
      >
        BUG BRAWL
      </motion.h1>

      <motion.input
        type="text"
        placeholder="Enter your name"
        className=" mb-6 px-4 py-2 rounded-full bg-[#9f9f9f] decoration-[#3a3a3a] font-bold text-2xl text-center"
        initial={{ opacity: 0, scale: 0 }}
        animate={{ opacity: 1, scale: 1, }}
        transition={{
          type: "spring",
          stiffness: 200,
          damping: 20,
        }}
      />

      <div className="flex gap-6 justify-evenly">
        {buttonLabels.map((label) => (
          <motion.button
            key={label}
            className="w-max lg:text-lg text-sm"
            initial={{ opacity: 0, scale: 0 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{
              type: "spring",
              stiffness: 200,
              damping: 20,
            }}
            whileHover={{ scale: 0.9 }}
            onClick={() => handleButtonClick(label)}
          >
            {label}
          </motion.button>
        ))}
      </div>
      <Popup show={showPopup} onClose={handleClosePopup} className="lg:w-1/4 lg:h-2/5">
        {selectedLabel === "Create Match" && (
          <div className="flex flex-col justify-between text-center items-center">
            <p>Enter Code</p>
            <motion.input
              type="text"
              className="bg-[#9f9f9f] rounded-lg m-3 px-4 py-1 w-2/3 text-black text-center !text-base"
            /><br />
            <motion.button
              className="!text-black w-1/3 !px-4 !py-2 !text-base !bottom-1"
              whileHover={{scale: 0.9}}

            >
              Create
            </motion.button>
          </div>
        )}
        {selectedLabel === "Join Match" && (
          <div className="flex flex-col justify-between text-center items-center">
            <p>Enter Code</p>
            <motion.input
              type="text"
              className="bg-[#9f9f9f] rounded-lg m-3 px-4 py-1 w-2/3 text-black text-center !text-base"
            /><br />
            <motion.button
              className="!text-black w-1/3 !px-4 !py-2 !text-base !bottom-1"
              whileHover={{scale:0.9}}
            >
              Join
            </motion.button>
          </div>
        )}
        {selectedLabel === "Find Match" && (
          <>
            <p>Finding Match</p>
            <p className="inline-block">
              {[0, 1, 2].map((i) => (
                <motion.span
                  key={i}
                  animate={{ y: [0, -5, 0] }}
                  transition={{
                    duration: 1.2,
                    repeat: Infinity,
                    ease: "easeInOut",
                    delay: i * 0.25,
                  }}
                  className="inline-block"
                >
                  .&nbsp;
                </motion.span>
              ))}
            </p>
          </>
        )}
      </Popup>
    </div>
  );
}
