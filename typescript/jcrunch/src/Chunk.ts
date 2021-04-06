const defaultInitialSize = 256;

export type Pointer = number;

export interface ChunkOpts {
  initialSize?: number;
  initialStorage?: ArrayBuffer;
}

export interface Reference {
  ptr: Pointer;
  buf: Uint8Array;
}

export default class Stack {
  private storage: ArrayBuffer;
  // sp acts as a top-of-stack pointer when Chunk is used as a stack
  // and a pointer to the next unallocated address when used as a heap.
  private sp: Pointer = 0;

  constructor(opts: ChunkOpts) {
    if (opts.initialStorage !== undefined) {
      this.storage = opts.initialStorage;
    } else if (opts.initialSize !== undefined) {
      this.storage = new ArrayBuffer(opts.initialSize);
    } else {
      this.storage = new ArrayBuffer(defaultInitialSize);
    }
  }

  pushUint8(n: number): Pointer {
    this.ensureCapacity(1);
    this.currentDataView().setUint8(0, n);
    return this.advance(1);
  }

  popUint8(): number {
    const val = this.currentDataView().getUint8(0);
    this.advance(1);
    return val;
  }

  pushUint16(n: number): Pointer {
    this.ensureCapacity(2);
    this.currentDataView().setUint16(0, n);
    return this.advance(2);
  }

  popUint16(): number {
    const val = this.currentDataView().getUint16(0);
    this.advance(2);
    return val;
  }

  pushUint32(n: number): Pointer {
    this.ensureCapacity(4);
    this.currentDataView().setUint32(0, n);
    return this.advance(4);
  }

  popUint32(): number {
    const val = this.currentDataView().getUint32(0);
    this.advance(4);
    return val;
  }

  pushInt8(n: number): Pointer {
    this.ensureCapacity(1);
    this.currentDataView().setInt8(0, n);
    return this.advance(1);
  }

  popInt8(): number {
    const val = this.currentDataView().getInt8(0);
    this.advance(1);
    return val;
  }

  pushInt16(n: number): Pointer {
    this.ensureCapacity(2);
    this.currentDataView().setInt16(0, n);
    return this.advance(2);
  }

  popInt16(): number {
    const val = this.currentDataView().getInt16(0);
    this.advance(2);
    return val;
  }

  pushInt32(n: number): Pointer {
    this.ensureCapacity(4);
    this.currentDataView().setInt32(0, n);
    return this.advance(4);
  }

  popInt32(): number {
    const val = this.currentDataView().getInt32(0);
    this.advance(4);
    return val;
  }

  alloc(size: number): Reference {
    this.ensureCapacity(size);
    const ptr = this.advance(size);
    return { ptr, buf: this.deref(ptr, size) };
  }

  deref(ptr: Pointer, size: number): Uint8Array {
    return new Uint8Array(this.storage, ptr, size);
  }

  setPtr(ptr: Pointer) {
    this.sp = ptr;
  }

  arrayBuffer(): ArrayBuffer {
    return this.storage.slice(0, this.sp);
  }

  rewind() {
    this.sp = 0;
  }

  private advance(bytes: number): Pointer {
    const ptr = this.sp;
    this.sp += bytes;
    return ptr;
  }

  private currentDataView(): DataView {
    return new DataView(this.storage, this.sp);
  }

  private ensureCapacity(extraBytes: number) {
    const currLen = this.storage.byteLength;
    if (currLen >= this.sp + extraBytes) {
      return;
    }

    const newBuf = new ArrayBuffer(currLen * 2);
    const copier = new Uint8Array(newBuf);
    copier.set(new Uint8Array(this.storage));

    this.storage = newBuf;
  }
}
