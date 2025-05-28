import React, { useState, useEffect, useRef } from 'react';
import { motion } from "framer-motion";
import { animate, eases } from 'animejs';
import bug2Gif from "../assets/image/bug2.gif";
import Bulb from "../assets/image/bulb.png";
import React, { useState, useEffect, useRef } from 'react';
import { motion } from "framer-motion";
import { animate, eases } from 'animejs';
import bug2Gif from "../assets/image/bug2.gif";
import Bulb from "../assets/image/bulb.png";

function Popup({ show, onClose, children, className = "", sabotageName = "" }) {
  const width = window.innerWidth;
  const height = window.innerHeight;
  const popupRef = useRef(null);
  const [bugs, setBugs] = useState([]);
  const [sabotage, setSabotage] = useState("");
  const [bugSwarmActive, setBugSwarmActive] = useState(false);
  const [bugLampActive, setBuglampActive] = useState(false);
  const [bugsFinished, setBugsFinished] = useState(0);
  const [bugEatActive, setBugEatActive] = useState(false);
  const [fakePopupActive, setFakePopupActive] = useState(false);
  const [fakePopups, setFakePopups] = useState([]);
  const [codeRainActive, setCodeRainActive] = useState(false);
  const [rainLine, setRainLine] = useState([]);
  const [scrambledText, setScrambledText] = useState(children);
  const [flickerActive, setFlickerActive] = useState(false);

  useEffect(() => {
    if (show && popupRef.current) {
      animate('popupRef.current', {
        opacity: [0, 1],
        scale: [0.9, 1],
        duration: 2000,
      });
    }
  }, [show]);

  useEffect(() => {
    switch (sabotageName) {
      case "BugSwarm":
        setBugSwarmActive(true);
        setSabotage("");
        break;
      case "BugLamp":
        setBuglampActive(true);
        setSabotage("");
        break;
      case "BugEat":
        setBugEatActive(true);
        setSabotage("");
        break;
      case "FakePopup":
        setFakePopupActive(true);
        setSabotage("");
        break;
      case "CodeRain":
        setCodeRainActive(true);
        setSabotage("");
        break;
      case "Flicker":
        setFlickerActive(true);
        break;
      default:
        setSabotage(sabotageName);
    }
  }, [sabotageName]);

  useEffect(() => {
    if (!bugSwarmActive) {
      setBugs([]);
      return;
    }

    const intervalId = setInterval(() => {
      setBugs(prev => {
        if (prev.length >= 400) return prev;


        const startX = Math.random() * width;
        const startY = Math.random() * height;
        const moveX = (Math.random() - 0.5) * 1500; // move in random X direction
        const moveY = (Math.random() - 0.5) * 1500; // move in random Y direction
        const rotate = Math.random() * 360;
        const newBug = {
          id: Date.now(),
          startX,
          startY,
          moveX,
          moveY,
          rotate,
        };
        return [...prev, newBug];
      });
    }, 50);

    return () => clearInterval(intervalId);
  }, [bugSwarmActive]);

  useEffect(() => {
    if (!bugLampActive) {
      setBugs([]);
      return;
    }

    const intervalId = setInterval(() => {
      setBugs(prev => {
        if (prev.length >= 20) return prev;

        const startX = (Math.random() * width) - 20;
        var startY;
        if (startX < -50 || startX > width) {
          startY = (Math.random() * height) + (height / 3);
        } else { startY = height }
        const moveX = (Math.random() * 20 + (width / 2 - 30));
        const moveY = (Math.random() * 50) + 10;
        var rotate;
        if (startX > width / 2) {
          rotate = -Math.tan((startX - width / 2) / startY)
        } else { rotate = Math.tan((width / 2 - startX) / startY) }
        const newBug = {
          id: Date.now(),
          startX,
          startY,
          moveX,
          moveY,
          rotate,
        };
        return [...prev, newBug];
      });
    }, 1200);

    return () => clearInterval(intervalId);
  }, [bugLampActive]);

  useEffect(() => {
    if (!fakePopupActive) {
      setFakePopups([]);
      return;
    }

    const totalPopups = 10;
    let count = 0;
    const interval = setInterval(() => {
      if (count >= totalPopups) {
        clearInterval(interval);
        return;
      }
      setFakePopups(prev => {
        const newPopup = {
          id: Date.now() + Math.random(),
          x: Math.random() * (window.innerWidth - 300),
          y: Math.random() * (window.innerHeight - 200),
        };
        count++;
        return [...prev, newPopup];
      });
    }, 1000);
    return () => clearInterval(interval);
  }, [fakePopupActive]);

  const handleCloseFake = (id) => {
    setFakePopups(prev => {
      const updated = prev.filter(p => p.id !== id);
      if (updated.length === 0) {
        setFakePopupActive(false);
      }
      return updated;
    });
  };

  function generateCodeLine() {
    const length = Math.floor(Math.random() * 20) + 10;
    let result = "";
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789{}();=+-*/<>";

    for (let i = 0; i < length; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }

    return result;
  }

  useEffect(() => {
    if (!codeRainActive) {
      setRainLine([]);
      return;
    }

    const createLine = () => {
      const length = Math.floor(Math.random() * 20) + 10;
      let codeLine = "";
      for (let i = 0; i < length; i++) {
        const charCode = Math.floor(Math.random() * (126 - 33 + 1)) + 33;
        codeLine += String.fromCharCode(charCode);
      }
      const newLine = {
        id: Date.now() + Math.random(),
        text: codeLine,
        left: Math.random() * width,
      };
      setRainLine(prev => [...prev, newLine]);
    };

    const interval = setInterval(() => {
      createLine();
    }, 20);

    return () => clearInterval(interval);
  }, [codeRainActive]);



  if (!show) return null;

  return (
    <div alt="popup" className={`${sabotage} fixed inset-0 bg-black/90 flex justify-center items-center z-50 `}
      style={{
        filter: `brightness(${Math.max(1 - bugsFinished * 0.05, 0.03)})`,
      }}
    >
      <div
        ref={popupRef}
        className={`${className} bg-[#3a3a3a] text-white border-t-40 border-[#9f9f9f] rounded-3xl shadow-lg 
        relative p-5 content-center text-center text-2xl font-bold`}
      >
        <motion.div
          className="absolute inset-0 w-full h-5 bg-white z-40"
          initial={{ width: '100%' }}
          animate={{ width: bugEatActive ? ['100%', '80%', '0%'] : '0%' }}
          transition={{ duration: bugEatActive ? 20 : 30, ease: 'linear', times: bugEatActive ? [0, 0.6, 1] : 1, }}
        />

        <motion.button
          className="absolute -top-10 right-4 !border-none !outline-none !bg-transparent text-black text-3xl !p-0"
          whileHover={{ scale: 0.9 }}

        >
          &times;
        </motion.button>
        {children}
      </div>

      {bugSwarmActive && bugs.map((bug) => (
        <motion.img
          key={bug.id}
          src={bug2Gif}
          alt="bug"
          className="absolute z-50 h-1/10"
          initial={{ left: `${bug.startX}px`, top: `${bug.startY}px` }}
          animate={{ x: `${bug.moveX}px`, y: `${bug.moveY}px`, rotate: `${bug.rotate}deg` }}
          transition={{ duration: 2, repeat: Infinity, repeatType: "reverse" }}
        />
      ))}

      {bugLampActive && (
        <>
          <img
            src={Bulb}
            className="absolute top-0 z-40 h-1/6"
          />
          {bugs.map((bug) => (
            <motion.img
              key={bug.id}
              src={bug2Gif}
              className="absolute z-50 h-1/15"
              initial={{ left: `${bug.startX}px`, top: `${bug.startY}px` }}
              animate={{ left: `${bug.moveX}px`, top: `${bug.moveY}px`, rotate: `${bug.rotate * (180 / Math.PI)}deg` }}
              transition={{ duration: 2 }}
              onAnimationComplete={() => setBugsFinished(prev => prev + 1)}
            />
          ))}
        </>
      )}

      {bugEatActive && (
        <motion.img
          src={bug2Gif}
          className="absolute -top-5 z-50 h-2/5 -rotate-90"
          initial={{ left: "100%" }}
          animate={{ left: ['100%', '70%', '-20%'] }}
          transition={{ duration: 25, ease: 'linear', times: [0, 0.4, 1], }}
        />
      )}

      {fakePopupActive && fakePopups.map(popup => (
        <div className="z-50">
          <motion.div
            key={popup.id}
            className="absolute bg-[#555555] text-white border-t-40 border-[#cccccc] rounded-3xl shadow-[0_10px_35px_black]
                     p-5 content-center text-center text-2xl font-bold lg:w-1/4 lg:h-2/5"
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            style={{ left: popup.x, top: popup.y }}
          >
            <motion.button
              className="absolute -top-10 right-4 !border-none !outline-none !bg-transparent text-black text-3xl !p-0"
              whileHover={{ scale: 0.9 }}
              onClick={() => handleCloseFake(popup.id)}
            >
              &times;
            </motion.button>
            <p>System Error !!!</p>
          </motion.div>
        </div>
      ))}

      {codeRainActive && rainLine.map((rain) => (
        <motion.div
          key={rain.id}
          className={`absolute text-green-500 font-bold text-2xl font-mono z-50 w-1 wrap-break-word`}
          initial={{ opacity: 0.2, left: rain.left, top: "-200%" }}
          animate={{ opacity: 1, top: "120%", }}
          transition={{ duration: 4, ease: "linear" }}
        >
          {rain.text}
        </motion.div>
      ))}
    </div>
  );
};

export default Popup;
