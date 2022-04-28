var ready = false;
var queue = [];
var inFlight = {};
var req_id = 0;

const worker = new window.Worker("worker.js");

worker.onmessage = (event) => {
    if (event.data && event.data.ready) {
        console.log("worker is ready");
        ready = true;
        for (var i = 0; i < queue.length; i++) {
            worker.postMessage(queue[i]);
        }
        return true;
    }
    console.log("EVENT");
    console.log(event.data);
    const waiting = inFlight[event.data.req_id];
    delete inFlight[event.data.req_id];
    if (!waiting) {
        console.log("got a response for an unregistered request: " + event.data.req_id);
        return true;
    }
    if (event.data.error) {
        console.log("rejecting");
        waiting.reject(event.data.error);
        return true;
    }
    console.log("resolving");
    waiting.resolve(event.data.payload);
};

console.log("loaded");

function sendMessage(data) {
    if (!data.req_id) {
        data.req_id = `req${++req_id}`;
    }
    console.log("DATA");
    console.log(data);
    const promise = new Promise((resolve, reject) => {
        inFlight[data.req_id] = {
            "resolve": resolve,
            "reject": reject,
        };
        // resolve("foo");
    })
    if (!ready) {
        console.log("not ready, queueing")
        queue.push(data);
        return promise;
    }
    worker.postMessage(data);
    return promise;
}

const keygen = async function keygen() {
    return sendMessage({ "action": "keygen" })
};

export { keygen };
