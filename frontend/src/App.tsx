import './App.css'
import "babel-polyfill";
import { useEffect, useState } from "react";
import SpeechRecognition, { useSpeechRecognition } from 'react-speech-recognition';
import { debounce } from 'underscore';
const synth = window.speechSynthesis;
// SpeechRecognition.startListening({ language: 'zh-CN' })

function App() {
    const {
        transcript,
        listening,
        resetTranscript,
        browserSupportsSpeechRecognition,
    } = useSpeechRecognition();
    const [gptResponse, setGptResponse] = useState('');
    const [loading, setLoading] = useState(false);
    const sendToChatGPT = async () => {
        setLoading(true);
        const res = await fetch(`http://localhost:8080/api/${transcript}`)
        const result = await res.json();
        const utterThis = new SpeechSynthesisUtterance(result.answer);
        utterThis.voice = synth.getVoices()[2];
        synth.speak(utterThis);
        setGptResponse(result.answer);
        setLoading(false);
    }
    const debouncedSendToChatGPT = debounce(sendToChatGPT, 100, false);
    if (!browserSupportsSpeechRecognition) {
        return <span>Browser doesn't support speech recognition.</span>;
    }

    useEffect(() => {
        if (!listening && transcript)
            debouncedSendToChatGPT();
    }, [listening])
    return (
        <div className="App">
            <div>
                <p>Microphone: {listening ? 'on' : 'off'}</p>
                <button onClick={() => SpeechRecognition.startListening({
                    language: 'ja'
                })}>Start</button>
                <button onClick={SpeechRecognition.stopListening}>Stop</button>
                <button onClick={resetTranscript}>Reset</button>
                <p>{transcript}</p>
                {loading ? <div className="loader">
                    <div className="inner one"></div>
                    <div className="inner two"></div>
                    <div className="inner three"></div>
                </div>
                    : ''}
                <p>{gptResponse}</p>
            </div>
        </div>
    )
}

export default App
