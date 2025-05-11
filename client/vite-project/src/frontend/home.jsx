import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import { useState } from "react";
import "../style/home.css"; 
import bugGif from "../assets/image/bug.gif";


function Home() {
  const [startTransition, setStartTransition] = useState(false);
  const navigate = useNavigate();

  return (
    <>
      <img src={bugGif} alt="Bug" className="w-max absolute top-1/2 left-full transform -translate-y-1/2 -rotate-90 scale-110 pointer-events-none fly-in" />

      {!startTransition && (
        <>
          <motion.h1 layoutId="name" className="left-full top-[30%] slide-from-right">BUG BRAWL</motion.h1>
          <motion.button 
          className="absolute w-max top-3/5 text-xl slide-from-right" 
          onClick={() => navigate("/lobby")}>
            Play Game
          </motion.button>
        </>
      )}
    </>
  );
};

export default Home;
