

class EuropaAudioProcessor extends AudioWorkletProcessor {
  constructor() {
    super();
  }

  process(_inputList: Float32Array[][], _outputList: Float32Array[][], _parameters: Record<string, unknown>) {
    /* using the inputs (or not, as needed), write the output
       into each of the outputs */


    return true;
  }
};

registerProcessor("europa-audio-processor", EuropaAudioProcessor);
