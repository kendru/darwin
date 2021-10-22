import { concert } from './notes';

const sampleRate = 48000;

interface DSPGenerator {
  getSample(n: number): number;
}

const dspGenFunc = (getSample: (n: number) => number): DSPGenerator => ({
  getSample,
});

const sineWave = (freq: number) => {
  const mult = (2 * Math.PI * freq) / sampleRate;
  return dspGenFunc((n) => Math.sin(n * mult));
};

const squareWave = (freq: number) => {
  const sine = sineWave(freq);
  return dspGenFunc((n) => Math.sign(sine.getSample(n)));
};

const add = (a: DSPGenerator, b: DSPGenerator) =>
  dspGenFunc((n) => a.getSample(n) + b.getSample(n));

const scaleConstant = (gen: DSPGenerator, alpha: number) =>
  dspGenFunc((n) => gen.getSample(n) * alpha);

const limit = (gen: DSPGenerator, min: number, max: number) => dspGenFunc((n) => Math.min(Math.max(gen.getSample(n), min), max));

const modulateAmplitude = (signal: DSPGenerator, lfo: DSPGenerator) =>
  dspGenFunc((n) => signal.getSample(n) * Math.abs(lfo.getSample(n)));

const durationSeconds = 5;

const audioCtx = new AudioContext({
  sampleRate,
});

const signalGen = limit(
  add(
    scaleConstant(sineWave(110), 0.5),
    modulateAmplitude(
      scaleConstant(squareWave(220), 0.3),
      sineWave(2),
    )
    ),
  -1,
  1
);


let startOffset = 0;

document.getElementById("play").addEventListener("click", () => {
  const ab = audioCtx.createBuffer(1, durationSeconds * sampleRate, sampleRate);

  for (let channel = 0; channel < ab.numberOfChannels; channel++) {
    const chBuf = ab.getChannelData(0);
    for (let n = 0; n < ab.length; n++) {
      const y = signalGen.getSample(n);
      chBuf[n] = y;
    }
  }

  const sourceNode = audioCtx.createBufferSource();
  sourceNode.buffer = ab;
  sourceNode.connect(audioCtx.destination);

  sourceNode.start(startOffset);
});

