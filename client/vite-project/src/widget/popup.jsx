import React, { useEffect, useRef } from 'react';
import {animate, eases} from 'animejs';

function Popup({ show, onClose, children }) {
  const popupRef = useRef(null);

  useEffect(() => {
    if (show && popupRef.current) {
      animate('popupRef.current',{
        opacity: [0, 1],
        scale: [0.9, 1],
        ease: eases.outExpo,
        duration: 400,
      });
    }
  }, [show]);

  if (!show) return null;

  return (
    <div className="fixed inset-0 bg-black/90 flex justify-center items-center z-50">
      <div
        ref={popupRef}
        className="bg-[#3a3a3a] text-white w-1/4 h-2/5 border-t-40 border-[#9f9f9f] rounded-3xl shadow-lg 
        relative p-5 content-center text-center text-2xl font-bold"
      >
        <button
          className="absolute -top-10 right-4 !bg-transparent text-black text-3xl !p-0"
          onClick={onClose}
        >
          &times;
        </button>
        {children} 
      </div>
  </div>
  );
};

export default Popup;
