var ready = false;
var queue = [];

const worker = new window.Worker("worker.js");

worker.onmessage = (event) => {
    if (event.data && event.data.ready) {
        console.log("worker is ready");
        ready = true;
        for (var i = 0; i < queue.length; i++) {
            worker.postMessage(queue[i]);
        }
        return;
    }
    console.log(event);
};

console.log("loaded");

function sendMessage(msg) {
    if (!ready) {
        console.log("not ready, queueing")
        queue.push(msg);
        return;
    }
    worker.postMessage(msg);
}

sendMessage("foo");
