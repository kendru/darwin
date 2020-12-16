import { aggregate, sum, namedNumberSum } from "./aggregate";

console.log(aggregate(sum, [1, 2, 3]));
console.log(aggregate(namedNumberSum, ["one", "two", "three"]));
