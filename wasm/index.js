const worker = new window.Worker("worker.js");

worker.onmessage = (event) => {
    console.log(event);
};

console.log("loaded");
