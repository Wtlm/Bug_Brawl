import React, { useState, useEffect } from 'react';
import '../style/sabotage.css';

export default function Sabotage({ isVisible, effectType }) {
  const [isSabotaging, setIsSabotaging] = useState(false);
  const [isFlickering, setIsFlickering] = useState(false);
  const [drift, setDrift] = useState({ x: 0, y: 0 });

  useEffect(() => {
    if (!isVisible) return;

    // Start flicker immediately when triggered
    setIsSabotaging(true);
    setIsFlickering(true);

    // Auto-disable after 3 seconds
    // const disableTimeout = setTimeout(() => {
    //   setIsSabotaging(false);
    //   setStaticIntensity(0);
    // }, 60000);

    // return () => {
    //   clearTimeout(disableTimeout);
    // };
  }, [isVisible]);

  // useEffect(() => {
  //   if (!isVisible || effectType !== 'mouse-drift') return;

  //   const handleMouseMove = () => {
  //     const offsetX = (Math.random() - 0.5) * 10; // ±5px
  //     const offsetY = (Math.random() - 0.5) * 10;
  //     setDrift({ x: offsetX, y: offsetY });
  //   };

  //   const interval = setInterval(() => {
  //     handleMouseMove();
  //   }, 100); // every 100ms

  //   const timeout = setTimeout(() => {
  //     clearInterval(interval);
  //     setDrift({ x: 0, y: 0 });
  //   }, 10000); // lasts for 10 seconds

  //   return () => {
  //     clearInterval(interval);
  //     clearTimeout(timeout);
  //   };
  // }, [isVisible, effectType]);

  // if (!isVisible) return null;

  return (
    <div
      className={`fixed top-0 left-0 w-screen h-screen z-50 pointer-events-none ${effectType}`}
      style={{
        transform: effectType === 'mouse-drift'
          ? `translate(${drift.x}px, ${drift.y}px)`
          : undefined,
        transition: 'transform 0.1s ease',
      }}
    >
    </div>
  );
}