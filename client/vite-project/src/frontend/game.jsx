import React, { useState, useEffect, useRef } from 'react';
import PlayCard from "./playcard";
import Popup from "../widget/popup";
import Questions from "../assets/quiz.json"
import { useLocation } from "react-router-dom";
import { getSocket } from "../socket/socket.js";
import { GameHandler } from "../socket/gameHandlers.js";
import { useNavigate } from 'react-router-dom';
import { useSocket } from '../socket/socketContext.jsx';



export default function Game() {
    const location = useLocation();
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
    const socket = useSocket();
    const [hasAnswered, setHasAnswered] = useState(false);
    const [selectedAnswerId, setSelectedAnswerId] = useState(null);
    const [waitWinner, setWaitWinner] = useState(null);

    console.log("Current Player ID:", currentPlayerId);
    console.log("Current Effects:", currentEffects);
    console.log("Sabotage Name:", sabotageName);


    // const [onGameOver, setOnGameOver] = useState(null);


    useEffect(() => {
        gameHandler.current = new GameHandler(socket, roomId, {
            setPlayers,
            setQuestion,
            setShowPopup,
            setTimeLeft,
            onTimeExpired: () => {
                console.log("Time's up!");
                gameHandler.current.sendAnswer("");
                setShowPopup(false);
                setQuestion(null);
                setCodeRainActive(false);
                setIsScrambled(false);
                setScrambledQuestion('');
                setScrambledOptions([]);
                // Add logic here to handle time expiry
            },
            setChooseSabotage,
            setWaitWinner,
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
            // socket.close();
        };
    }, [roomId]);

    useEffect(() => {
        if (currentEffects.includes("CodeRain")) {
            setCodeRainActive(true);
        } else {
            setCodeRainActive(false);
        }

        const questionTime = currentEffects.includes("BugEat") ? 20 : 5;
        gameHandler.current.startTimer(questionTime);
    }, [currentEffects]);

    useEffect(() => {
        // Hide question popup if any other popup is active
        if (!!chooseSabotage || !!sabotageNoti || !!roundResultNoti || waitWinner) {
            setShowPopup(false);
        }
    }, [chooseSabotage, sabotageNoti, roundResultNoti, waitWinner]);

    useEffect(() => {
        // Reset answer state only when a new question arrives
        if (question) {

            setChooseSabotage(null);
            setSabotageNoti(null);
            setRoundResultNoti(null);
            setWaitWinner(null);
            setHasAnswered(false);
            setSelectedAnswerId(null);
            setShowPopup(true);
        }
    }, [question]);

    useEffect(() => {
        // Reset answer state only when sabotage selection appears
        if (!!chooseSabotage) {
            setHasAnswered(false);
            setSelectedAnswerId(null);
        }
    }, [chooseSabotage]);

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
    };

    useEffect(() => {
        if (currentEffects.includes("CodeRain")) {
            setCodeRainActive(true);
        } else {
            setCodeRainActive(false);
        }

        if (currentEffects.includes("BugEat")) {
            // trigger BugEat effect here (e.g., set a state or call a function)
            // Example: bugEatActiveRef.current = true;
        } else {
            // reset BugEat effect if needed
            // Example: bugEatActiveRef.current = false;
        }
    }, [currentEffects]);

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
            <Popup className="w-3/5 h-5/6" show={showPopup} sabotageName={sabotageName}>
                {question && (
                    <div className={` p-6 text-left  ${isScrambled ? 'text-green-500' : 'text-white'}`}>
                        <h2 className="lg:text-2xl text-sm font-bold mb-5 "> {isScrambled ? scrambledQuestion : question.question}</h2>
                        <ul className=" grid grid-cols-2 grid-rows-2 gap-5">
                            {(isScrambled ? scrambledOptions : question.options.map(o => o.text)).map((text, idx) => {
                                const optionId = question.options[idx].id;
                                const isSelected = selectedAnswerId === optionId;

                                return (
                                    <li
                                        key={optionId}
                                        className={`${isScrambled ? 'bg-black' : isSelected ? 'bg-white' : 'bg-[#9f9f9f]'} ${isScrambled ? 'text-green-400' : 'text-black'} px-4 py-1 rounded-xl font-medium lg:text-xl text-xs content-center leading-tight
                                                     ${hasAnswered ? 'opacity-60 cursor-not-allowed' : 'hover:bg-white cursor-pointer'}`}
                                        onClick={() => {
                                            if (hasAnswered) return;
                                            setHasAnswered(true);
                                            setSelectedAnswerId(optionId);
                                            gameHandler.current.sendAnswer(optionId);
                                        }}
                                    >
                                        {/* <span className="mr-2">{optionId.toUpperCase()}.</span> */}
                                        {text}
                                    </li>
                                );
                            })}

                        </ul>
                    </div>
                )}
            </Popup>
            <Popup
                className="w-3/5 h-5/6"
                show={!!chooseSabotage || !!sabotageNoti || !!roundResultNoti || !!waitWinner}
                sabotageName=""
            >
                {sabotageNoti ? (
                    <div className="p-6 text-center text-white lg:text-2xl text-sm font-bold mb-5">
                        {sabotageNoti}
                    </div>
                ) : roundResultNoti ? (
                    <div className="p-6 text-center text-white lg:text-2xl text-sm font-bold mb-5">
                        {roundResultNoti}
                    </div>
                ) : waitWinner ? (
                    <div className="p-6 text-center text-white lg:text-2xl text-sm font-bold mb-5">
                        Waiting for {waitWinner} to choose a sabotage...
                    </div>
                ) : chooseSabotage ? (
                    <div className="p-6 text-center text-white">
                        <h2 className="lg:text-2xl text-sm font-bold mb-5">Choose a sabotage</h2>

                        <ul className="grid grid-cols-2 grid-rows-2 gap-5">
                            {chooseSabotage.map(choice => {
                                const isSelected = selectedAnswerId === choice;
                                return (
                                    <li
                                        key={choice}
                                        className={`${isSelected ? 'bg-white' : 'bg-[#9f9f9f]'} text-black px-4 py-1 rounded-xl font-medium lg:text-xl text-xs content-center leading-tight cursor-pointer 
                                                    ${hasAnswered ? 'opacity-60 cursor-not-allowed' : 'hover:bg-white cursor-pointer'}`}
                                        onClick={() => {
                                            if (hasAnswered) return;
                                            setHasAnswered(true);
                                            setSelectedAnswerId(choice);
                                            gameHandler.current.sendSabotageChoice(choice);

                                        }}
                                    >
                                        {choice}
                                    </li>
                                );
                            })}
                        </ul>
                    </div>
                ) : null}
            </Popup>


        </div>
    );
}
