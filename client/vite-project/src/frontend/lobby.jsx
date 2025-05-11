import React, { useState } from 'react';
import { delay, motion } from "framer-motion";
import "../style/lobby.css";
import Popup from "../widget/popup";

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

  return (
    <>
      <motion.h1 
        layoutId="name" 
        className="top-[10%]"
        initial={{ scale: 1 }}
        animate={{ scale: 0.7}}
        transition={{ type: "spring", duration: 1}}
      >
        BUG BRAWL
      </motion.h1>

      <motion.input
        type="text"
        placeholder="Enter your name"
        className="absolute top-[48%] left-1/2 px-4 py-2 rounded-full bg-[#9f9f9f] decoration-[#3a3a3a] font-bold text-xl text-center"
        initial={{ opacity: 0, scale: 0 }}
        animate={{ opacity: 1, scale: 1, x: "-50%", y: "-50%",}}
        transition={{ 
          type: "spring",
          stiffness: 200,
          damping: 20, }}
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
            whileHover={{scale: 0.9}}
            onClick={() => handleButtonClick(label)}
          >
            {label}
          </motion.button>
        ))}
      </div>
      <Popup show={showPopup} onClose={handleClosePopup}>
        {selectedLabel === "Create Match" && (
          <div className="flex flex-col justify-between text-center items-center">
            <p>Enter Code</p>
            <motion.input
              type="text"
              className="bg-[#9f9f9f] rounded-lg m-3 px-4 py-1 w-2/3 text-black text-center !text-base"
            /><br/>
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
            /><br/>
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
    </>
  );
}
