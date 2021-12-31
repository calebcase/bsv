// Package decimal provides a fixed point base 10 number.
//
// The equation for a decimal number is:
//
//  number = value * 10 ^ scale
//
// Where number is fixed point number, value is an unscaled integer, and scale
// is base 10 exponent. For example:
//
//  1.23 = 123 * 10^-2
//
// Scale may be up to ±2^21 (approximately a decimal number with 2 million
// zeros). Value may be up to ±2^34_359_738_343 (approximately
// ±10^10_343_311_884).
//
// Encoding
//
// The decimal is laid out first by the unscaled integer value (with sign bit),
// then scale value (with sign bit), and finally the last 2 bits are the scale
// size.
//
// Decoding is expected to use the BSV block to first read in the full data,
// discover the scale size from the last two bits, extract the scale (up to 3
// bytes total), and then the remaining bits are the value.
//
// Encoding will first determine the size of the scale needed and then
// determine the smallest control block size that can represent it and the
// value.
//
// All integers in the format are encoded big-endian with a trailing sign bit
// (aka zigzag).
//
// The scale size is encoded as two bits:
//
//  | 0 | 1 | Available Scale |
//  |-------|-----------------|
//  | 0 . 0 | No Scale        | 1 byte, remaining bits in this byte are used for value.
//  | 0 . 1 | ±2^5 Scale      | 1 byte, remaining bits are the scale value.
//  | 1 . 0 | ±2^12 Scale     | 2 bytes
//  | 1 . 1 | ±2^21 Scale     | 3 bytes
//  |-------|-----------------|
//  | 0 | 1 |
//
// Delta Encoding
//
// Delta encoded decimals restrict the use of scale to Data Size and Data Size
// Size control blocks. Data, Data+1 and Data+2 blocks implicitly use the most
// recently specified scale. The intention of this arrangement is to maximize
// the data bits available when delta encoding (given that delta encoding is
// expected to produce small values with similar scales).
//
// Examples
//
// Small Values No Precision (1 byte)
//
//  2^4 - 1 = ±15
//
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//  |---------------|---------------|
//  | 1 | 0 . 0 . 0 . 0 | 0 | 0 . 0 | Data Control Block with value of 0.
//  |---------------|---------------|
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//
// USD 0.0001 (2 bytes)
//
//  2^4 - 1 = ±15
//
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//  |---------------|---------------|
//  | 0 . 0 . 1 | 0 . 0 . 0 . 1 | 0 | Data + 1 Control Block with value of +1.
//  |-------------------------------|
//  | 0 . 0 . 1 . 0 . 0 | 1 | 0 . 1 | ±2^5 Scale with scale of -4.
//  |---------------|---------------|
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//
// USD 20.47 (3 bytes)
//
//   2^11 - 1 = ±2_047
//
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//  |---------------|---------------|
//  | 0 . 0 . 0 . 1 | 1 . 1 . 1 . 1 | Data + 2 Control Block with value of +2047.
//  | 1 . 1 . 1 . 1 . 1 . 1 . 1 | 0 |
//  |-------------------------------|
//  | 0 . 0 . 0 . 1 . 0 | 1 | 0 . 1 | ±2^5 Scale with scale of -2.
//  |---------------|---------------|
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//
// USD 3.2767 (4 bytes)
//
//  2^15 - 1 = ±32_767
//
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//  |---------------|---------------|
//  | 0 . 1 | 0 . 0 . 0 . 0 . 1 . 1 | Data Size Control Block with value of 3.
//  |-------------------------------|
//  | 1 . 1 . 1 . 1 . 1 . 1 . 1 . 1 | Value of +32767
//  | 1 . 1 . 1 . 1 . 1 . 1 . 1 | 0 |
//  |-------------------------------|
//  | 0 . 0 . 1 . 0 . 0 | 1 | 0 . 1 | ±2^5 Scale with scale of -4.
//  |---------------|---------------|
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//
// USD 1677.7215 (5 bytes)
//
//  2^23 - 1 = ±16_777_215
//
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//  |---------------|---------------|
//  | 0 . 1 | 0 . 0 . 0 . 1 . 0 . 0 | Data Size Control Block with value of 4.
//  |---------------|---------------|
//  | 1 . 1 . 1 . 1 . 1 . 1 . 1 . 1 | Value of +16777215
//  | 1 . 1 . 1 . 1 . 1 . 1 . 1 . 1 |
//  | 1 . 1 . 1 . 1 . 1 . 1 . 1 | 0 |
//  |---------------|---------------|
//  | 0 . 0 . 1 . 0 . 0 | 1 | 0 . 1 | ±2^5 Scale with scale of -4.
//  |---------------|---------------|
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//
// Ethereum 1 Wei
//
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//  |---------------|---------------|
//  | 0 . 1 | 0 . 0 . 0 . 0 . 1 | 0 | Data + 1 Control Block with value of +1.
//  |---------------|---------------|
//  | 1 . 0 . 0 . 1 . 0 | 1 | 0 . 1 | ±2^5 Scale with scale of -18.
//  |---------------|---------------|
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
//
package decimal
