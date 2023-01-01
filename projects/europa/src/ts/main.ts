interface EuropaAudioContext {
  audioContext: AudioContext,
  audioProcessor: AudioWorkletNode,
}

async function main() {
  async function initEuropaAudio(): Promise<EuropaAudioContext|null> {
    let audioContext = null;
    try {
      audioContext = new AudioContext();
      await audioContext.resume();
      await audioContext.audioWorklet.addModule("/assets/js/audio-worklet.js");
    } catch(e) {
      return null;
    }

    const audioProcessor = new AudioWorkletNode(audioContext, "europa-audio-processor");

    return {
      audioContext,
      audioProcessor,
    };
  }

  let ctx = await initEuropaAudio();
  console.log('Started', ctx);
}
(async () => {
  const td = new TextDecoder();
  const mem = new WebAssembly.Memory({ initial: 2, maximum: 2 });
  const env = {
    log: (ptr: number, len: number) => {
      const strSlice = mem.buffer.slice(ptr, ptr + len);
      const str = td.decode(strSlice);
      console.log({ ptr, len, str, strSlice, memSize: mem.buffer.byteLength });
    },
  };

  const mod = await WebAssembly.instantiateStreaming(fetch('/assets/wasm/xyzzy.wasm'), {
    env,
    js: { mem },
  });
  const instance = mod.instance;
  const result = instance.exports.add(42, 8);
  console.log('Result from Zig:', result);
})();
main();
