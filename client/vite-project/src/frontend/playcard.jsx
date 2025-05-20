import React from 'react';
import gearPic from "../assets/image/gear.png"

const borderColors = {
  blue: 'border-blue-500',
  red: 'border-red-500',
  green: 'border-green-500',
  yellow: 'border-yellow-400',
};

export default function PlayCard({ name, health = 3, borderColor = 'blue' }) {
  const hearts = Array.from({ length: 5 }, (_, i) => (
    <span key={i}>
      {i < health ? '❤️' : '🤍'}
    </span>
  ));

  return (
    <div className={`relative border-4 ${borderColors[borderColor]} h-full`}>
      <h2 className="absolute w-full text-white lg:text-base sm:text-xs text-[8px] font-bold z-50 px-3">{name}</h2>
      <div className="relative w-full h-full flex justify-center">
        <img src={gearPic} alt="Computer" className="w-full h-full object-fill" />
        <div className="absolute lg:text-2xl sm:text-base text-[8px] top-1/3 text-center">{hearts}</div>
      </div>
    </div>
  );
}
