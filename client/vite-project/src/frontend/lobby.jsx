import React, { useState } from 'react';
import { delay, motion } from "framer-motion";
import "../style/lobby.css";
import Popup from "../widget/popup";

import Sabotage from '../widget/sabotage.jsx';

export default function Lobby() {
  const buttonLabels = ["Create Match", "Join Match", "Find Match"];
  const [selectedLabel, setSelectedLabel] = useState('');
  const [showPopup, setShowPopup] = useState(false);


  const handleButtonClick = (label) => {
    setSelectedLabel(label);
    setShowPopup(true);
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
    //<div className={showGlitch ? "font-glitch" : ""}></div>
    //<div className={isBackwards ? "backwards-text" : ""}></div>
    <div
      className="drift-effect"
      style={{
        transform: `translate(${drift.x}px, ${drift.y}px)`
      }}>
      <motion.h1
        layoutId="name"
        className="top-[10%]"
        initial={{ scale: 1 }}
        animate={{ scale: 0.7 }}
        transition={{ type: "spring", duration: 1 }}
      >
        BUG BRAWL
      </motion.h1>

      <motion.input
        type="text"
        placeholder="Enter your name"
        className="absolute top-[48%] left-1/2 px-4 py-2 rounded-full bg-[#9f9f9f] decoration-[#3a3a3a] font-bold text-xl text-center"
        initial={{ opacity: 0, scale: 0 }}
        animate={{ opacity: 1, scale: 1, x: "-50%", y: "-50%", }}
        transition={{
          type: "spring",
          stiffness: 200,
          damping: 20,
        }}
      />

      <div className="absolute flex gap-6 justify-evenly top-[62%] left-1/2 transform -translate-1/2">
        {buttonLabels.map((label) => (
          <motion.button
            key={label}
            className="text-lg"
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

      <div className="flex flex-col justify-between text-center items-center">
        <div className="absolute top-[75%] left-[30%] transform -translate-x-1/2">
          <motion.button
            className="!text-black w-full !px-4 !py-2 !text-base !bottom-1"
            onClick={handleFlickerClick}>
            Test Effect Flicker
          </motion.button>
        </div>
        <div className="absolute top-[75%] left-[45%] transform -translate-x-1/2">
          <motion.button
            className="!text-black w-full !px-4 !py-2 !text-base !bottom-1"
            onClick={handleBlurryClick}>
            Test Effect Blurry
          </motion.button>
        </div>
        <div className="absolute top-[75%] left-[60%] transform -translate-x-1/2">
          <motion.button
            className="!text-black w-full !px-4 !py-2 !text-base !bottom-1"
            onClick={handleGlitchClick}>
            Test Font Glitch
          </motion.button>
        </div>
        <div className="absolute top-[75%] left-[75%] transform -translate-x-1/2">
          <motion.button
            className="!text-black w-full !px-4 !py-2 !text-base !bottom-1"
            onClick={handleBackwardsClick}>
            Test Backwards Text
          </motion.button>
        </div>
        <div className="absolute top-[75%] left-[90%] transform -translate-x-1/2">
          <motion.button
            className="!text-black w-full !px-4 !py-2 !text-base"
            onClick={handleMouseDriftClick}>
            Test Mouse Drift
          </motion.button>
        </div>
      </div>

      {/* Broken Screen Effect */}
      <Sabotage isVisible={showFlicker} effectType={'screen-flicker-overlay'} />
      <Sabotage isVisible={showBlurry} effectType={'backdrop-blur-[2px]'} />

      <Popup show={showPopup} onClose={handleClosePopup}>
        {selectedLabel === "Create Match" && (
          <div className="flex flex-col justify-between text-center items-center">
            <p>Enter Code</p>
            <motion.input
              type="text"
              className="bg-[#9f9f9f] rounded-lg m-3 px-4 py-1 w-2/3 text-black text-center !text-base"
            /><br />
            <motion.button
              className="!text-black w-1/3 !px-4 !py-2 !text-base !bottom-1">
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
              className="!text-black w-1/3 !px-4 !py-2 !text-base !bottom-1">
              Join
            </motion.button>
          </div>
        )}
        {selectedLabel === "Find Match" && (
          <>
            <p>Finding Match</p>
            <p className="dot">
              <span>.&nbsp;</span>
              <span>.&nbsp;</span>
              <span>.</span>
            </p>
          </>
        )}
      </Popup>
    </div>
  );
}
