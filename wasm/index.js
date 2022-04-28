import { keygen, build } from "./gobl.js";

const result = await keygen();
console.log(`RESULT: ${result}`);

try {
    const buildResult = await build({
        "data": {},
        "privateykey": {},
    });
    console.log(`BUILD RESULT: ${build_result}`);
} catch (e) {
    console.log("BUILD ERROR: " + e)
};

const result2 = await keygen();
console.log(`RESULT2: ${result2}`);
