// [
//     [state, updater],
//     ...
// ]
const hookSlots = [];
let i = 0;

function useState(initialValue) {
  const j = i;
  if (hookSlots[j] === undefined) {
    hookSlots[j] = [
      initialValue,
      (newValue) => {
        hookSlots[j][0] = newValue;
      },
    ];
  }
  i++;

  return hookSlots[j];
}

const effectStack = [];
function useEffect(fn) {
  const j = i;
  if (hookSlots[j] === undefined) {
    hookSlots[j] = fn;
    effectStack.push(j);
  }
  i++;
}

const registeredFunctions = [];
const registerFunction = (fn) => {
  registeredFunctions.push(fn);
};

registerFunction(() => {
  const [count, setCount] = useState(0);
  setCount(count + 1);

  return count;
});

registerFunction(() => {
  const [count, setCount] = useState(0);

  setCount(count + 2);

  return count % 100 === 0 ? "at one hundred" : "-";
});

const prevRuns = [];
function run() {
  i = 0;

  for (const fnIdx in registeredFunctions) {
    const fn = registeredFunctions[fnIdx];
    const res = fn();
    if (prevRuns[fnIdx] !== undefined && prevRuns[fnIdx] != res) {
      console.log(`Got new result: ${res}`);
    }
    prevRuns[fnIdx] = res;
  }

  setTimeout(run, 50);
}
run();

for (const effectIdx of effectStack) {
  const effect = hookSlots[effectIdx];
  effect();
}
