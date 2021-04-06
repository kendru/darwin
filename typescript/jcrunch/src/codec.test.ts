import { assert } from "console";
import { encode, decode } from "./codec";

const tests = [
  { name: "null", in: null },
  { name: "number: 0", in: 0 },
  { name: "number: positive", in: 1 },
  { name: "number: negative", in: -1 },
  { name: "string: empty", in: "" },
  // { name: "string: ascii", in: "taco friday" },
  // { name: "string: unicode", in: "🌮 friday! 🎉" },
  // { name: "empty array", in: [] },
  // { name: "numeric array", in: [1, 2, 3] },
  // { name: "string array", in: ["", "a", "b", "c"] },
  // { name: "empty object", in: {} },
  // { name: "object: homogenous map", in: { a: 12, b: 13, c: 14 } },
  // { name: "object: heterogenous map", in: { name: "Charles", toes: 10 } },
  // {
  //   name: "complex nested object",
  //   in: {
  //     name: "Andrew",
  //     age: 31,
  //     hobbies: ["jazz piano", "baking", "kettlebells"],
  //     children: [
  //       {
  //         name: "Audrey",
  //         age: 8,
  //         hobbies: ["reading", "dolls"],
  //       },
  //       {
  //         name: "Jonah",
  //         age: 6,
  //         hobbies: ["costumes", "running"],
  //       },
  //     ],
  //   },
  // },
];

tests.forEach((t) =>
  test(t.name, () => {
    const encoded = encode(t.in);
    const out = decode(encoded);
    expect(out).toEqual(t.in);
    // console.log({
    //   test: t.name,
    //   stringSize: JSON.stringify(t.in).length * 2,
    //   encodedSize: encoded.byteLength,
    // });
  })
);
