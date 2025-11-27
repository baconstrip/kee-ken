import mitt from 'mitt';

type Events = {
    "beginCountdown": string,
    "countdownStart"?: null,
    "submitBid": number,
    "submitAnswer": string,
    "spacePress"?: null,
    "editorAuthSuccess"?: null,
    "buzz": number,
    "cancelGame"?: null,
    "selectQuestion": string,
    "nextRound"?: null,
    "markCorrect"?: null,
    "markIncorrect"?: null,
    "finishReading"?: null,
    "moveOn"?: null,
    "authReady"?: null,
    "hideQuestion"?: null,
};

const eventBus = mitt<Events>();

export default eventBus;