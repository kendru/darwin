const crypto = require('crypto');

const charTbl = '0123456789abcdefghjkmnpqrstvwxyz';
const encodeChunk = (n) => charTbl[n];

const nthChunk = (input, idx) => {
	const startBit = 5*idx;
	const offset = startBit % 8;
	const startByte = (startBit - offset)/8;

	const mask = 0xff >> offset;
	let n = input[startByte] & mask;
	const postShift = 3-offset;
	if (postShift < 0) {
		n = n << -postShift;
		
		let nextByte = 0;
		if (input.length > startByte+1) {
			nextByte = input[startByte+1];
		}
		const secondByteBits = nextByte >> (8+postShift);
		n |= secondByteBits;
	} else {
		n = n >> postShift;
	}

	return n;
}

const encodeBase32 = (input) => {
	let out = '';
	
	const totalBytes = input.length * 8;
	for (let i = 0; i*5 <= totalBytes; i++) {
		out += encodeChunk(nthChunk(input, i))
	}

	return out;
};

// Node.JS compatible.
const uniqueId = () => encodeBase32(crypto.randomBytes(16))

// WebCrypto compatible.
// const uniqueId = () => encodeBase32(crypto.getRandomValues(new Uint8Array(new ArrayBuffer(16))));

const main = async () => {
	for (let index = 0; index < 100; index++) {
		console.log(uniqueId());
	}
}
main();
