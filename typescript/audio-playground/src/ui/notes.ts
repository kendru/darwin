// Mapping of MIDI note number to frequency.
type Scale = (midiNum: number) => number;

export const concert: Scale = (midiNum: number): number => 440*Math.pow(2, (midiNum-69)/12);
