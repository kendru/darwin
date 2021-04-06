export enum Opcode {
  ReadNull = 0,

  // Read a numeric value.
  ReadUint8,
  ReadUint16,
  ReadUint32,
  ReadInt8,
  ReadInt16,
  ReadInt32,

  // Load value from offset in data stack.
  Load8,
  Load16,
  Load32,
}
