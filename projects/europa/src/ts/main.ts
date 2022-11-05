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

  let ctx = initEuropaAudio();
  console.log('Started', ctx);
}
main();
