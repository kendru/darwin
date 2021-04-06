import * as fs from "fs/promises";
import { decode, encode } from "./codec";
import { TextEncoder, TextDecoder } from "util";
import { crunch, uncrunch } from "graphql-crunch";
import vm from "./vm";

const exit = (msg: string, exitCode: number = 0) => {
  console.error(msg);
  process.exit(exitCode);
};

const mustTry = async (fn: () => Promise<void>) => {
  try {
    return await fn();
  } catch (err) {
    if (err.code == "ENOINT") {
      return exit(`File not found: ${(err as any).path}`);
    }

    return exit(`Quit with error: ${err.message}`, 1);
  }
};

async function main() {
  const [_interp, _script, filename] = process.argv;
  const json = await mustTry(async () => {
    const f = await fs.readFile(filename, "utf-8");
    return JSON.parse(f);
  });

  console.log("============== DEBUG ==============");
  let v = vm({});
  v.writeObj(json);
  v.debug();
  const encoded = v.state();

  // const v1 = vm({ initialState: encoded });
  // console.log("======= Decoding vm =======");
  // v1.debug();
  // const encoded = encode(json);

  const jsonLen = tmpJSLength(json);
  const gqlCrunchLen = tmpJSLength(crunch(json));
  const jCrunchLen = encoded.byteLength;
  console.log(`JSON\t(${jsonLen} bytes)`);
  console.log(`GQL\t(${gqlCrunchLen} bytes)`);
  console.log(`JCRUNCH\t(${jCrunchLen} bytes)`);
  console.log(
    `RATE/JSON:\t${(100 - (encoded.byteLength / jsonLen) * 100).toFixed(2)}%`
  );
  console.log(
    `RATE/GQL:\t${(100 - (encoded.byteLength / gqlCrunchLen) * 100).toFixed(
      2
    )}%`
  );
  crunch();

  // console.log("Decoded:", decode(encoded));
}

function tmpJSLength(o: any): number {
  const encodedJSON = JSON.stringify(o);
  return new TextEncoder().encode(encodedJSON).byteLength;
}

main();
