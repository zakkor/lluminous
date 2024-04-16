export async function compressAndEncode(messages) {
  const json = JSON.stringify(messages);
  const encoder = new TextEncoder();
  const uint8Array = encoder.encode(json);

  const compressedStream = new window.CompressionStream('deflate');
  const compressed = new Blob([uint8Array]).stream().pipeThrough(compressedStream);
  const compressedArrayBuffer = await new Response(compressed).arrayBuffer();
  const compressedUint8Array = new Uint8Array(compressedArrayBuffer);
  return btoa(String.fromCharCode(...compressedUint8Array));
}

export async function decodeAndDecompress(encoded) {
  const binaryString = atob(encoded);
  const len = binaryString.length;
  const uint8Array = new Uint8Array(new ArrayBuffer(len));
  for (let i = 0; i < len; i++) {
      uint8Array[i] = binaryString.charCodeAt(i);
  }

  const decompressedStream = new DecompressionStream('deflate');
  const decompressed = new Blob([uint8Array]).stream().pipeThrough(decompressedStream);
  const decompressedArrayBuffer = await new Response(decompressed).arrayBuffer();
  const decoder = new TextDecoder();
  return JSON.parse(decoder.decode(decompressedArrayBuffer));
}