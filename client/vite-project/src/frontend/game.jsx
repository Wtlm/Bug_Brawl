import React, { useState, useEffect } from 'react';
import PlayCard from "./playcard";
import Popup from "../widget/popup";
import Questions from "../assets/quiz.json"


export default function Game() {
    const [showPopup, setShowPopup] = useState(false);
    const [question, setQuestion] = useState(null);
    const [isScrambled, setIsScrambled] = useState(false);
    const [scrambledQuestion, setScrambledQuestion] = useState('');
    const [scrambledOptions, setScrambledOptions] = useState([]);


    useEffect(() => {
        const timer = setTimeout(() => {
            const randomIndex = Math.floor(Math.random() * Questions.length);
            setQuestion(Questions[randomIndex]);
            setShowPopup(true);
            setCodeRainActive(true)
        }, 2000);
        // Cleanup if component unmounts before timeout
        return () => clearTimeout(timer);
    }, []);

    useEffect(() => {
        if (!showPopup || !question) return;

        const scrambleText = (text) => {
            if (!text) return "";
            const chars = text.split('');
            const numToScramble = Math.min(4, chars.length);
            for (let i = 0; i < numToScramble; i++) {
                const index = Math.floor(Math.random() * chars.length);
                const randChar = String.fromCharCode(Math.floor(Math.random() * (126 - 33 + 1)) + 33);
                chars[index] = randChar;
            }
            return chars.join('');
        };

        setScrambledQuestion(scrambleText(question.question));
        setScrambledOptions(question.options.map(opt => scrambleText(opt.text)));

        // Toggle isScrambled every 5 seconds
        const interval = setInterval(() => {
            setIsScrambled(prev => !prev);
        }, 3000);

        return () => clearInterval(interval);
    }, [question, showPopup]);


    const handleClosePopup = () => {
        setShowPopup(false);
    };

    return (
        <div className="h-screen bg-black grid grid-cols-2 grid-rows-2 justify-center items-center gap-3 p-3">
            <PlayCard name="Chanh" health={5} borderColor="blue" />
            <PlayCard name="HaDo" health={2} borderColor="red" />
            <PlayCard name="Sieu Nhan Long Long" health={3} borderColor="green" />
            <PlayCard name="Sieu Nhan Long Long" health={4} borderColor="yellow" />


            <Popup className="w-3/5 h-5/6" show={showPopup} onClose={handleClosePopup} sabotageName="CodeRain">
                {question && (
                    <div className={` p-6 text-left  ${isScrambled ? 'text-green-500' : 'text-white'}`}>
                        <h2 className="lg:text-2xl text-sm font-bold mb-5 "> {isScrambled ? scrambledQuestion : question.question}</h2>
                        <ul className=" grid grid-cols-2 grid-rows-2 gap-5">
                            {(isScrambled ? scrambledOptions : question.options.map(o => o.text)).map((text, idx) => (
                                <li
                                    key={question.options[idx].id}
                                    className={`${isScrambled ? 'bg-black' : 'bg-[#9f9f9f]'} ${isScrambled ? 'text-green-400' : 'text-black'} px-4 py-1 rounded-xl font-medium lg:text-xl text-xs content-center leading-tight`}
                                >
                                    {/* <span className="mr-2">{option.id.toUpperCase()}.</span> */}
                                    {text}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
            </Popup>
        </div>
    );
}
