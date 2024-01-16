import './App.css'
import "babel-polyfill";
import { useEffect, useState } from "react";
import SpeechRecognition, { useSpeechRecognition } from 'react-speech-recognition';
import { debounce } from 'underscore';
const synth = window.speechSynthesis;

// Tabs
import { Tab, Tabs, TabList, TabPanel } from 'react-tabs';
import 'react-tabs/style/react-tabs.css';
// SpeechRecognition.startListening({ language: 'zh-CN' })

import ReactMarkdown from 'react-markdown'
import remarkMath from 'remark-math'
import rehypeKatex from 'rehype-katex'
import 'katex/dist/katex.min.css'

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
        const res = await fetch(`http://localhost:8081/api/${transcript}`)
        const result = await res.json();
        const utterThis = new SpeechSynthesisUtterance(result.answer);
        utterThis.voice = synth.getVoices()[2];
        synth.speak(utterThis);
        if (result.answer != null) {
            setGptResponse(result.answer);
        }
        setLoading(false);
    }

    const [text, setText] = useState<string>('');
    const [history, setHistory] = useState<string[]>([]);
    const [content, setContent] = useState<string>('');
    const [loading2, setLoading2] = useState(false);
    const sendToAPI = async () => {
        setLoading2(true);
        console.log(`http://localhost:8081/api/${text}`);

        const res = await fetch(`http://localhost:8081/api/${text}`)
        const result = await res.json();

        console.log(result.answer);
        if (result.answer != null) {
            const tempText = "#### " + text + "\n:  " + result.answer;
            setHistory((prev) => [...prev, tempText])


            const tempText2 = history.join()
            setContent((prev) => prev + "\n\n" + tempText)
            setText("")
        }


        setLoading2(false);
    }


    const handleChangeText = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setText(e.target.value);
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

        <Tabs className="App">
            <TabList >
                <Tab>Voice</Tab>
                <Tab>Text</Tab>
            </TabList>

            <TabPanel>
                <p>Mic: {listening ? 'on' : 'off'}</p>
                <button onClick={() => SpeechRecognition.startListening({
                    language: 'ja'
                })}>Start</button>
                <button onClick={SpeechRecognition.stopListening}>Stop</button>
                <button onClick={resetTranscript}>Reset</button>
                <p>{transcript}</p>
                {loading ?
                    <div className="loader">
                        <div className="inner one"></div>
                        <div className="inner two"></div>
                        <div className="inner three"></div>
                    </div>
                    : ''}
                <ReactMarkdown
                    rehypePlugins={[rehypeKatex]}
                    remarkPlugins={[remarkMath]}>
                    {gptResponse}
                </ReactMarkdown>
            </TabPanel>
            <TabPanel>
                {loading2 ?
                    <div className="loader">
                        <div className="inner one"></div>
                        <div className="inner two"></div>
                        <div className="inner three"></div>
                    </div>
                    : ''}
                <ReactMarkdown
                    rehypePlugins={[rehypeKatex]}
                    remarkPlugins={[remarkMath]}>
                    {content}
                </ReactMarkdown>

                <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
                    <textarea rows={5} cols={50} value={text} onChange={handleChangeText} />
                    <input type="submit" style={{ textAlign: "right" }} onClick={sendToAPI} value="Send" />
                </div>

            </TabPanel>
        </Tabs>
    )
}

export default App
