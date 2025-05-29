import React, { useState, useEffect, useRef } from 'react';
import PlayCard from "./playcard";
import Popup from "../widget/popup";
import Questions from "../assets/quiz.json"
import { useLocation } from "react-router-dom";
import { getSocket } from "../socket/socket.js";
import { GameHandler } from "../socket/gameHandlers.js";
import { useNavigate } from 'react-router-dom';



export default function Game() {
    const location = useLocation();

    console.log("Game component mounted:", location.state?.roomId);
    const [showPopup, setShowPopup] = useState(false);
    const [question, setQuestion] = useState(null);
    const [isScrambled, setIsScrambled] = useState(false);
    const [scrambledQuestion, setScrambledQuestion] = useState('');
    const [scrambledOptions, setScrambledOptions] = useState([]);
    const [codeRainActive, setCodeRainActive] = useState(false);
    const gameHandler = useRef(null);
    const [players, setPlayers] = useState(location.state?.players || []);
    const [roomId, setRoomId] = useState(location.state?.roomId || null);
    const [timeLeft, setTimeLeft] = useState(30);
    const bugEatActiveRef = useRef(false);
    const [chooseSabotage, setChooseSabotage] = useState(null);
    const [sabotageNoti, setSabotageNoti] = useState(null);
    const [playerEffects, setPlayerEffects] = useState({});
    const currentPlayerId = location.state?.playerId;
    const currentEffects = playerEffects[currentPlayerId] || [];
    const sabotageName = currentEffects.length > 0 ? currentEffects[0] : "";
    const [roundResultNoti, setRoundResultNoti] = useState(null);
    const navigate = useNavigate();

    // const [onGameOver, setOnGameOver] = useState(null);


    useEffect(() => {
        console.log("Game component mounted");
        const socket = getSocket();
        // console.log("Game component mounted with roomId:", roomId);
        // console.log("players:", players);
        gameHandler.current = new GameHandler(socket, roomId, {
            setPlayers,
            setQuestion,
            setShowPopup,
            setTimeLeft,
            onTimeExpired: () => {
                console.log("Time's up!");
                setShowPopup(false);
                // Add logic here to handle time expiry
            },
            setChooseSabotage,
            setCodeRainActive,
            setSabotageNoti,
            setPlayerEffects,
            setRoundResultNoti,
            onGameOver: () => {
                console.log("Game Over!");
                navigate("/gameover");
            }
        });

        // Cleanup if needed
        return () => {
            gameHandler.current?.clearTimer?.();
            socket.close();
        };
    }, [roomId]);


    // useEffect(() => {
    //     const timer = setTimeout(() => {
    //         const randomIndex = Math.floor(Math.random() * Questions.length);
    //         setQuestion(Questions[randomIndex]);
    //         setShowPopup(true);
    //         // setCodeRainActive(true)
    //     }, 2000);
    //     // Cleanup if component unmounts before timeout
    //     return () => clearTimeout(timer);
    // }, [roomId]);

    useEffect(() => {
        if (!showPopup || !question || !codeRainActive) return;

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
    }, [question, showPopup, codeRainActive]);


    const handleClosePopup = () => {
        setShowPopup(false);
        setCodeRainActive(false);
        setIsScrambled(false);
        setScrambledQuestion('');
        setScrambledOptions([]);
    };

    return (

        <div className="h-screen bg-black grid grid-cols-2 grid-rows-2 justify-center items-center gap-3 p-3">

            {players.map((player, idx) => (
                <PlayCard
                    key={player.id || idx}
                    name={player.name}
                    id={player.id}
                    health={player.health || 5} // default to 5 if not sent
                    borderColor={["blue", "red", "green", "yellow"][idx % 4]}
                />
            ))}
            <Popup className="w-3/5 h-5/6" show={showPopup} onClose={handleClosePopup} sabotageName={sabotageName}>
                {question && (
                    <div className={` p-6 text-left  ${isScrambled ? 'text-green-500' : 'text-white'}`}>
                        <h2 className="lg:text-2xl text-sm font-bold mb-5 "> {isScrambled ? scrambledQuestion : question.question}</h2>
                        <ul className=" grid grid-cols-2 grid-rows-2 gap-5">
                            {(isScrambled ? scrambledOptions : question.options.map(o => o.text)).map((text, idx) => (
                                <li
                                    key={question.options[idx].id}
                                    className={`${isScrambled ? 'bg-black' : 'bg-[#9f9f9f]'} ${isScrambled ? 'text-green-400' : 'text-black'} px-4 py-1 rounded-xl font-medium lg:text-xl text-xs content-center leading-tight`}
                                    onClick={() => {
                                        gameHandler.current.sendAnswer(question.options[idx].id);
                                    }}
                                >
                                    {/* <span className="mr-2">{option.id.toUpperCase()}.</span> */}
                                    {text}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
            </Popup>
            <Popup
                className="w-3/5 h-5/6"
                show={!!chooseSabotage}
                sabotageName=""
            >
                {sabotageNoti && (
                    <div className=" p-6 text-center text-white lg:text-2xl text-sm font-bold mb-5">
                        {sabotageNoti}
                    </div>
                )}
                {roundResultNoti && (
                    <div className=" p-6 text-center text-white lg:text-2xl text-sm font-bold mb-5">
                        {roundResultNoti}
                    </div>
                )}
                {chooseSabotage && (
                    <div className="p-6 text-center text-white">
                        <h2 className="lg:text-2xl text-sm font-bold mb-5">
                            Choose a sabotage
                        </h2>

                        <ul className="grid grid-cols-2 grid-rows-2 gap-5">
                            {chooseSabotage.choices.map(choice => (
                                <li
                                    key={choice}
                                    className="bg-[#9f9f9f] text-black px-4 py-1 rounded-xl font-medium lg:text-xl text-xs content-center leading-tight cursor-pointer"
                                    onClick={() => {
                                        chooseSabotage.choose(choice);
                                        setChooseSabotage(null);
                                    }}
                                >
                                    {choice}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
            </Popup>

        </div>
    );
}
