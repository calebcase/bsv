// Package control provides the fundamental BSV blocking structure.
//
// BSV control blocks use a prefix coding scheme to indicate the type of the
// current byte (which then further indicates how many bytes the field
// contains). The intention is to minimize signaling overhead and pack as much
// data directly into the control block as possible.
//
// Control Block
//
// This diagram indicates the bits that are fixed (filled in) vs bits that are
// available for encoding data (blanks). Data and size information is expected
// to be extracted by masking off the fixed bits. This is only the first byte
// (several control block types are multi-byte sequences).
//
//  | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 || Type           |                                                                            |
//  |---------------|---------------||----------------|----------------------------------------------------------------------------|
//  | 1 |                           || Data           | 2^7 = 128 values                                                           |
//  | 0 . 1 |                       || Data Size      | 2^6 = 64 bytes; 2^(64*8) values                                            |
//  | 0 . 0 . 1 |                   || Data + 1       | 2^(5+8) = 2^13 = 8192 values                                               |
//  | 0 . 0 . 0 . 1 |               || Data + 2       | 2^(4+8+8) = 2^20 = 1048576 values                                          |
//  | 0 . 0 . 0 . 0 . 1 |           || Skip           | 2^3 = 8 fields                                                             |
//  | 0 . 0 . 0 . 0 . 0 . 1 |       || Data Size Size | 2^2 = 4 bytes size; 2^(4*8) = 2^32 = 4294967296 bytes; 2^4294967296 values |
//  | 0 . 0 . 0 . 0 . 0 . 0 . 1 |   || Skip Size      | 2^1 = 2 bytes size; 2^(2*8) = 2^16 = 65536 fields                          |
//  | 0 . 0 . 0 . 0 . 0 . 0 . 0 . 1 || Null           | Null value (for nullable fields)                                           |
//  | 0 . 0 . 0 . 0 . 0 . 0 . 0 . 0 || Empty          | Empty value                                                                |
//  |---------------|---------------||----------------|----------------------------------------------------------------------------|
//
// All sizes and skips are indexed starting at 1 to maximize their effective
// range. To encode zero length data (e.g. empty string) use the Empty block.
//
// Data blocks allow for 7 bits of data to be encoded directly into the block.
// Encoding up to 128 values (e.g. integers between -63 and +63).
//
// Data Size blocks have two parts:
//
//  1. Number of bytes that contain data
//  2. Data
//
// Data Size blocks allow for up to 64 bytes of data. They are meant to
// efficiently encode short data sequences like short strings or medium sized
// numbers.
//
// Data + 1 blocks are two byte sequences containing data. They allow for 13
// bits of data to be encoded with up to 8192 values (e.g. integers between
// -4095 and +4095).
//
// Data + 2 blocks are three byte sequences containing data. They allow for 20
// bits of data to be encoded with up to 1048576 values (e.g. integers between
// -524287 and +524287).
//
// Skip blocks indicate that up to 8 contiguous fields aren't directly encoded
// in the byte stream.  The specific meaning of what the skipped fields contain
// is schema and encoding dependent. For example, a delta encoder might
// interpret default to mean "whatever the previous value was" rather than
// specifically an all zero byte sequence. This is meant to efficiently encode
// short runs of sparse data (e.g. delta encoded values that aren't changing
// over several rows and ).
//
// Data Size Size blocks have 3 parts:
//
//  1. Number of bytes for the data size
//  2. Number of bytes that contain data
//  3. Data
//
// Data Size Size blocks allow for up to 4294967296 bytes of data (4 GiB).
//
// Skip Size blocks have two parts:
//
//  1. Number of bytes that contain skip
//  2. Skip
//
// Skip Size blocks indicate that up to 65536 contiguous fields are their
// default value and aren't directly encoded in the byte stream. This is meant
// to efficiently encode very sparse data streams (e.g. large sparse matrices
// with many zeros).
//
// Null blocks indicate that the field is set to the null value (e.g. a string
// field that is optional, but needs to differentiate between an empty string
// and no value provided).
//
// Empty blocks indicate that the field is set to the empty value (e.g. an
// empty string). Empty blocks may take on additional meaning by schema
// processors. For example they may be used to indicate the end of an unbounded
// repeating field (e.g. array of integers).
package control
