import { TextEncoder, TextDecoder } from "util";
import { Opcode } from "./opcode";
import Chunk, { Pointer } from "./Chunk";

export enum ValueType {
  Null = 0,
  Number,
  String,
  ArrayStart,
  ArrayEnd,
  ObjectStart,
  ObjectEnd,
}

export interface VM {
  writeObj(obj: any);
  read(): unknown;
  rewind();
  state(): ArrayBuffer;

  // Prints state.
  debug();
}

export interface VMOpts {
  initialState?: ArrayBuffer;
}

const parseHeader = (
  state: ArrayBuffer
): { instrLen: number; dataLen: number } => {
  const dv = new DataView(state);
  const instrLen = dv.getUint32(0);
  const dataLen = dv.getUint32(4);

  return {
    instrLen,
    dataLen,
  };
};

const writeHeader = (state: ArrayBuffer, instrLen: number, dataLen: number) => {
  const dv = new DataView(state);
  dv.setUint32(0, instrLen);
  dv.setUint32(4, dataLen);
};

const dumpBuffer = (buf: ArrayBuffer) => {
  let line: number[] = [];
  let i = 0;

  const asHexByte = (b: number) => b.toString(16).padStart(2, "0");
  const asAscii = (b: number) => {
    const str = String.fromCharCode(b);
    if (b == 0) {
      return ".";
    }
    if (b < 20 || b > 126) {
      return "�";
    }
    return str;
  };
  const printLine = () =>
    console.log(
      "\t" + line.map(asHexByte).join(" ") + " | " + line.map(asAscii).join("")
    );

  while (i < buf.byteLength) {
    const dv = new DataView(buf, i);
    const b = dv.getUint8(0);
    line.push(b);
    if ((i + 1) % 32 == 0) {
      printLine();
      line = [];
    }
    i++;
  }

  if (line.length > 0) {
    printLine();
  }
};

const vm = (opts: VMOpts): VM => {
  // The instructions keeps the operations for building an object.
  let instr: Chunk;
  // The data stack holds values that are pointed to by references in the stack.
  let data: Stack;

  // This contains a map of values that we have seen before to their location in the
  // `data` heap. Before we write a value, we check to see if we can replace it with a
  // pointer.
  const offsets: Map<string, Pointer> = new Map();

  if (opts.initialState) {
    const headerLen = 8;
    const { instrLen, dataLen } = parseHeader(opts.initialState);
    const stackEnd = headerLen + instrLen;
    instr = new Chunk({
      initialStorage: opts.initialState.slice(headerLen, stackEnd),
    });
    instr.setPtr(instrLen);
    data = new Stack({
      initialStorage: opts.initialState.slice(stackEnd, stackEnd + dataLen),
    });
    data.setPtr(dataLen);
  } else {
    instr = new Chunk({});
    data = new Stack({});
  }
  // The general strategy is to write each value in-place but keep track of what has
  // been written so that when we see a previously-seen value, we write a pointer instead
  // (unless the value itself is smaller than a pointer).
  // Additionally, the stack/heap as storage containers is the wrong abstraction.
  // Instead, we should have a "data" stack where we store _all_ values (not just
  // variable-length data) and an "instructions" stack. As we encode a value, we
  // pop an instruction onto the  instruction stack and data onto the data stack.
  // We need to define the opcodes for this VM.
  // This allows us to more compactly represent things like homogenous arrays because
  // we could generate something like the following:
  // Data: 1 4 95 6 23
  // Instructions:
  // BEGIN_ARRAY
  // INIT_LOOP (sets global loop counter to 0)
  // label: loop1
  // READ_NUM
  // JUMP_LOOP_LT 5 @loop1
  // END_ARRAY
  //
  // instead of:
  // BEGIN_ARRAY
  // READ_NUM
  // READ_NUM
  // READ_NUM
  // READ_NUM
  // READ_NUM
  // END_ARRAY
  //
  // Additionally, we can implement reads of objects that share structure similar to how we would implement
  // functions:
  // label: decode_obj1
  // BEGIN_OBJ
  // READ_KEY
  // READ_STRING
  // BEGIN_OBJ
  // JMP ; jump to return address

  // const readString = (): string => {
  //   sp -= 1;
  //   const tag = readTag();
  //   let ptr: number;
  //   if (tag.hasFlag(ValueFlag.Reference8)) {
  //     ptr = popByte();
  //   } else if (tag.hasFlag(ValueFlag.Reference16)) {
  //     ptr = readIUit16();
  //   } else if (tag.hasFlag(ValueFlag.Reference32)) {
  //     ptr = readIUit32();
  //   } else {
  //     throw new Error("Expected string to store pointer");
  //   }
  //   const lenBuf = deref(ptr, 4);
  //   const len = new DataView(
  //     lenBuf.buffer,
  //     lenBuf.byteLength,
  //     lenBuf.byteLength
  //   ).getUint32(0);
  //   const text = deref(ptr + 4, len);
  //   const td = new TextDecoder();
  //   return td.decode(text);
  // };

  // const writeTag = (tag: ValueTag) => pushByte(tag.byte());
  // const readTag = (): ValueTag => ValueTag.parse(popByte());
  // const peekTag = (): ValueTag => {
  //   const t = ValueTag.parse(popByte());
  //   sp -= 1;
  //   return t;
  // };

  const pushOpcode = (opcode: Opcode) => {
    stack.pushUint8(opcode as number);
  };

  const popOpcode = (): Opcode => stack.popUint8() as Opcode;

  const writeNull = () => pushOpcode(Opcode.ReadNull);

  const writeNumber = (n: number) => {
    if (n < -32768) {
      pushOpcode(Opcode.ReadInt32);
      data.pushInt32(n);
      return;
    }

    if (n < -128) {
      pushOpcode(Opcode.ReadInt16);
      data.pushInt16(n);
      return;
    }

    if (n < 0) {
      pushOpcode(Opcode.ReadInt8);
      data.pushInt8(n);
      return;
    }

    if (n < 256) {
      pushOpcode(Opcode.ReadUint8);
      data.pushUint8(n);
      return;
    }

    if (n < 65536) {
      pushOpcode(Opcode.ReadUint16);
      data.pushUint16(n);
      return;
    }

    pushOpcode(Opcode.ReadUint32);
    data.pushUint32(n);
  };

  // Save string to heap, push a load of the pointer to the string, then push a read string to the stack.
  const writeString = (s: string) => {
    let ptr = offsets.get(s);
    if (ptr === undefined) {
      const te = new TextEncoder();
      const bytes = te.encode(s);
      const strLen = bytes.length;
      const bufLen = 4 + strLen;
      const { ptr: allocPtr, buf } = data.alloc(bufLen);
      ptr = allocPtr;
      // Buffer contains string length followed by bytes of the string
      // TODO: Consider variable-length tag.
      const dv = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
      dv.setUint32(0, strLen);
      buf.set(bytes, 4);
      offsets.set(s, ptr);
    }

    // Write pointer
    if (ptr <= 255) {
      pushOpcode(Opcode.Load8);
      stack.pushUint8(ptr);
    } else if (ptr <= 65535) {
      pushOpcode(Opcode.Load16);
      stack.pushUint16(ptr);
    } else {
      pushOpcode(Opcode.Load32);
      stack.pushUint32(ptr);
    }
  };

  const write = (val: any) => {
    switch (typeof val) {
      case "number":
        writeNumber(val);
        return;

      case "string":
        writeString(val);
        return;

      case "object":
        if (val === null) {
          writeNull();
          return;
        }

      // if (Array.isArray(object)) {
      //   write(null, ValueType.ArrayStart);
      //   for (const elem of object) {
      //     this.writeObj(elem);
      //   }
      //   write(null, ValueType.ArrayEnd);
      //   break;
      // }

      // // TODO: Consider other possible cases of object.

      // write(null, ValueType.ObjectStart);
      // for (const key in object) {
      //   if (object.hasOwnProperty(key)) {
      //     this.writeObj(key);
      //     this.writeObj(object[key]);
      //   }
      // }
      // write(null, ValueType.ObjectEnd);
      // break;

      default:
        throw new Error(`Cannot handle object type: ${typeof val}`);
    }
  };

  return {
    writeObj(object: any) {
      write(object);
    },

    rewind() {
      stack.rewind();
      data.rewind();
    },

    read(): unknown {
      const opcode = popOpcode();
      switch (opcode) {
        case Opcode.ReadNull:
          return null;

        case Opcode.ReadUint8:
          return data.popUint8();

        case Opcode.ReadUint16:
          return data.popUint16();

        case Opcode.ReadUint32:
          return data.popUint32();

        case Opcode.ReadInt8:
          return data.popInt8();

        case Opcode.ReadInt16:
          return data.popInt16();

        case Opcode.ReadInt32:
          return data.popInt32();

        // case Opcode.ReadString:
        //   return readString();

        // case ValueType.ArrayStart:
        //   // TODO: should we write the array length?
        //   const v = [];
        //   while (peekTag().valueType != ValueType.ArrayEnd) {
        //     v.push(this.read());
        //   }
        //   readTag();
        //   return v;
        // case ValueType.ObjectStart: {
        //   const v = {};
        //   while (peekTag().valueType != ValueType.ObjectEnd) {
        //     const key = this.read();
        //     const val = this.read();
        //     v[key] = val;
        //   }
        //   readTag();
        //   return v;
        // }
        // default:
        //   break;
      }
    },

    debug() {
      // console.log(`Stack Size: ${sp} bytes`);
      // console.log(`Data Size: ${dp} bytes`);
      // console.log("Stack:");
      // dumpBuffer(stack);
      // console.log("Data:");
      // dumpBuffer(data);
    },

    state(): ArrayBuffer {
      const usedStack = stack.arrayBuffer();
      const usedData = data.arrayBuffer();
      const state = new ArrayBuffer(
        8 + usedStack.byteLength + usedData.byteLength
      );

      writeHeader(state, usedStack.byteLength, usedData.byteLength);
      const copier = new Uint8Array(state, 8);
      copier.set(new Uint8Array(usedStack), 0);
      copier.set(new Uint8Array(usedData), usedStack.byteLength);

      return state;
    },
  };
};

export default vm;
