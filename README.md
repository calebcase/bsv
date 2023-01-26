# BSV

BSV [Control Blocks](#control-blocks) use a prefix coding scheme to indicate
the type of the current byte (which then further indicates how many bytes the
field contains). The intention is to minimize signaling overhead and pack as
much data directly into the control block as possible.

BSV does not specify the contents of the data bits. The meaning of the data
bits must come from some additional schema or specification supplied to the
encoder/decoder.

## Control Blocks

This diagram indicates the bits that are fixed (filled in) vs bits that are
available for encoding data (blanks). Data and size information is expected to
be extracted by masking off the fixed bits. This is only the first byte
(several control block types are multi-byte sequences).

| `\| 0123 4567 \|` | Type                                  | Abbr           | Min Bytes | Max Bytes    |
|-------------------|---------------------------------------|---------------:|----------:|-------------:|
| `\| 1... .... \|` | [Data](#data)                         | `d`            | 1         | 1            |
| `\| 01.. .... \|` | [Data Size](#data-size)               | `dz`           | 2         | 65           |
| `\| 001. .... \|` | [Data + 1](#data-%2B-1)               | `d1`           | 2         | 2            |
| `\| 0001 .... \|` | [Data + 2](#data-%2B-2)               | `d2`           | 3         | 3            |
| `\| 0000 1... \|` | [Data Size Size](#data-size-size)     | `dzz`          | 3         | 9 + $2^{64}$ |
| `\| 0000 01.. \|` | [Container Blocks](#container-blocks) | `cb` `cu` `ce` | 2         | ∞            |
| `\| 0000 001. \|` | [Skip Size](#skip-size)               | `sz`           | 2         | 3            |
| `\| 0000 0001 \|` | [Empty](#empty)                       | `e`            | 1         | 1            |
| `\| 0000 0000 \|` | [Null](#null)                         | `n`            | 1         | 1            |

Data and size embedded in the control block is extracted by masking all control
bits to zero. For example `1|000_0001` decodes to `0000_0001` and `001|1_0001
0000_0000` decodes to `0001_0001 0000_0000`.

Some of the types have overlapping ranges. For example, 8-bits of data can be
stored in either of `dz`, `d1`, `d2`, or `dzz` blocks. Encoders optimizing
overhead will prefer the most compact type (`d1`), but they are not required
to. This allows encoders to make trade offs for different optimization choices
like maximizing encoding speed or minimizing mutable field causing rewrites of
all following data.

The blocking structure has the following absolute and efficient data to total
bytes ratios. The absolute values are a direct consequence of the entire
encodable range for that block type. The efficient values are what you would
expect if picking the block type with the lowest overhead for the amount of
data. For example, data with 14 to 20 bits would be stored more efficiently in
a `d2` block rather than a `dz` or `dzz` block.

| Type  | Abs. Data        | Abs. Data %    | Eff. Data         | Eff. Data %    |
|------:|-----------------:|---------------:|------------------:|---------------:|
| `d`   | 7 bits           | 88.5%          | 7 bits            | 88.5%          |
| `dz`  | 1-64 bytes       | 50% to 98.4%   | 3-64 bytes        | 75% to 98.4%   |
| `d1`  | 13 bits          | 81.3%          | 13 bits           | 81.3%          |
| `d2`  | 20 bits          | 83.3%          | 20 bits           | 83.3%          |
| `dzz` | 1-$2^{64}$ bytes | 33.3% to ~100% | 65-$2^{64}$ bytes | 97.0% to ~100% |

All sizes are indexed starting at 1 to maximize their effective range. To
encode zero length data (e.g. empty string) use the [Empty](#empty) block. For
example, a size of 1 in a [Data Size](#data-size) block is encoded as
`01|00_0000` and size 2 as `01|00_0001`.

### Data

|              |                    |
|--------------|--------------------|
| Abbreviation | `d`                |
| Capacity     | $2^7$ = 128 values |

Data blocks allow for 7 bits of data to be encoded directly into the block.
Encoding up to 128 values (e.g. booleans, signed integers between -64 and +64).

```
.               .
|0 1 2 3 4 5 6 7|
+-+-+-+-+-+-+-+-+
|1|    data     |
+-+-+-+-+-+-+-+-+
```

```
.               .
|0 1 2 3 4 5 6 7|
+-+-+-+-+-+-+-+-+
|1|    data     |
+-+-+-+-+-+-+-+-+
```

For example, the following could represent boolean values:

```
.               .
|0 1 2 3 4 5 6 7|
+-+-+-+-+-+-+-+-+
|1|0 0 0 0 0 0 0|
+-+-+-+-+-+-+-+-+
        ^
        |
        * boolean false
```

```
.               .
|0 1 2 3 4 5 6 7|
+-+-+-+-+-+-+-+-+
|1|0 0 0 0 0 0 1|
+-+-+-+-+-+-+-+-+
        ^
        |
        * boolean true
```

### Data Size

|              |                                         |
|--------------|-----------------------------------------|
| Abbreviation | `dz`                                    |
| Capacity     | $2^6$ = 1 to 64 bytes; $2^{512}$ values |

Data Size blocks have two parts:

1. Number of bytes that contain data
2. Data

Data Size blocks allow for up to 64 bytes of data. They are meant to
efficiently encode short data sequences like short strings or medium sized
numbers.

```
.0              .    1         1.               .6           7  .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |4 5 6 7 8 9 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
|0 1|   size    |                1-64 bytes data                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
```

Examples:

```
.0              .    1          .        2      .            3  .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|4 5 6 7 8 9 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 1|0 0 0 0 1 0|                 3 bytes data                  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
      size = 3
```

```
.0              .    1         1.               .6           7  .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |4 5 6 7 8 9 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
|0 1|1 1 1 1 1 1|                 64 bytes data                 |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
      size = 64
```

### Data + 1

|              |                                    |
|--------------|------------------------------------|
| Abbreviation | `d1`                               |
| Capacity     | $2^{5+8}$ = $2^{13}$ = 8,192 bytes |

Data + 1 blocks are two byte sequences containing data. They allow for 13 bits
of data to be encoded with up to 8,192 values (e.g. signed integers between
-4,096 and +4,096).

```
.0              .    1          .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 1|       13 bits data      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### Data + 2

|              |                                          |
|--------------|------------------------------------------|
| Abbreviation | `d2`                                     |
| Capacity     | $2^{4+8+8}$ = $2^{20}$ = 1,048,576 bytes |

Data + 2 blocks are three byte sequences containing data. They allow for 20
bits of data to be encoded with up to 1,048,576 values (e.g. integers between
-524,288 and +524,288).

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 1|               20 bits data            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### Data Size Size

|              |                                                          |
|--------------|----------------------------------------------------------|
| Abbreviation | `dzz`                                                    |
| Capacity     | $2^3$ = 1 to 8 bytes size; $2^{8*8}$ = $2^{64}$ = 16 EiB |

Data Size Size blocks have 3 parts:

1. Number of bytes for the data size
2. Number of bytes that contain data
3. Data

Data Size Size blocks allow for up to 16 EiB (exbibytes) of data.

```
.0              .    1         1.               .               .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |               |               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++
|0 0 0 0 1|  Z  |         1-8 bytes size        |         1-16 EiB data         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++
Z: bytes of size
```

Examples:

```
                                                                 2     2       2
                                                                 0     0       0
.0              .    1          .        2     2.               .4     5       5.
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|               |7 8 9 0 1 2 3 5|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
|0 0 0 0 1|0 0 0|1 1 1 1 1 1 1 1|                256 bytes data                 |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
bytes of size = 1
size = 256
```

### Container Blocks

Container blocks signal the start and end of an embedded BSV. There are two
types of container blocks for known size (bounded) and unknown size (unbounded)
situations.

| `\| 0123 4567 \|` | Type                                        | Abbr |
|-------------------|---------------------------------------------|-----:|
| `\| 0000 0111 \|` | Reserved                                    |      |
| `\| 0000 0101 \|` | [Container Bounded](#container-bounded)     | `cb` |
| `\| 0000 0110 \|` | [Container Unbounded](#container-unbounded) | `cu` |
| `\| 0000 0100 \|` | [Container End](#container-unbounded)       | `ce` |

#### Container Bounded

|              |                                                                                           |
|--------------|-------------------------------------------------------------------------------------------|
| Abbreviation | `cb`                                                                                      |
| Capacity     | $2^{(2^{64}*8)}$ = $2^{(2^{64}*2^3)}$ = $2^{2^{(64+3)}}$ = $2^{2^{68}}$ = $2^{136}$ bytes |

Bounded containers have 3 parts:

1. Control Block Container Bounded
2. Size encoded in a single BSV field
3. An embedded BSV

The largest bounded container would use a size encoded in a `dzz` block
with $2^{64}*8$ bits and an embedded BSV of $2^{(2^{64}*8)}$ bytes.

```
.0              .    1         1.               .               .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |               |               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++
|0 0 0 0 0 1|0 1|       size as BSV field       |              BSV              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++
        ^
        |
        * `cb` block
```

Bounded containers can be efficiently seeked past if the embedded data isn't
needed immediately.

A bounded block with an embedded BSV (e.g. a struct with a single boolean field
set to true):

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|1|0 0 0 0 0 0 0|1|0 0 0 0 0 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        ^               ^               ^
        |               |               |
        |               |               * boolean true in a `d` block
        |               |
        |               * 1 byte size in a `d` block
        |
        * `cb` block
```

A bounded block that is empty (e.g. an empty struct):

```
.0              .    1         1.
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|0 0 0 0 0 0 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        ^               ^
        |               |
        |               * `e` block
        |
        * `cb` block
```

This is equivalent to the longer form:

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|1|0 0 0 0 0 0 0|0 0 0 0 0 0 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        ^               ^               ^
        |               |               |
        |               |               * `e` block
        |               |
        |               * 1 byte size in a `d` block
        |
        * `cb` block
```

A bounded block that is null (e.g. a pointer to a struct):

```
.0              .    1          .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|0 0 0 0 0 0 0 0|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        ^               ^
        |               |
        |               * `n` block
        |
        * `cb` block
```

This is also equivalent to the longer form:

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|1|0 0 0 0 0 0 0|0 0 0 0 0 0 0 0|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        ^               ^               ^
        |               |               |
        |               |               * `n` block
        |               |
        |               * 1 byte size in a `d` block
        |
        * `cb` block
```

NOTE: Depending on the semantic needs of the encoder/decoder and the
capabilities of the schema it may be sufficient to simply use the Empty and
Null blocks directly (instead of embedding them in a Container block).

#### Container Unbounded

|              |           |
|--------------|-----------|
| Abbreviation | `cb` `ce` |
| Capacity     | ∞         |

`cu` `ce`

Unbounded containers blocks have 3 parts:

1. Control Block Container Being Unbounded
2. An embedded BSV
3. Control Block Container End

Processing unbounded blocks requires counting the open and closing blocks to
ensure they are properly paired. Seeking past the data requires scanning the
bytes to find the closing block.

```
.0              .    1         1.               .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|1 0|              BSV              |0 0 0 0 0 1|0 0|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
        ^                                               ^
        |                                               |
        * `cu` block                                    * `ce` block
```

An unbounded block with an embedded BSV (e.g. an array of integers):

```
.0              .    1          .        2      .            3  .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|4 5 6 7 8 9 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|1 0|1|0 0 0 0 0 0 0|1|0 0 0 0 0 0 1|0 0 0 0 0 1|0 0|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        ^        \                             /        ^
        |         -----------------------------         |
        |                       ^                       * `ce` block
        |                       |
        |                       * embedded BSV with two fields
        |
        * `cu` block
```

### Skip Size

|              |                                                              |
|--------------|--------------------------------------------------------------|
| Abbreviation | `sz`                                                         |
| Capacity     | $2^1$ = 1 to 2 bytes; $2^{(2*8)}$ = $2^{16}$ = 65,536 fields |

Skip Size blocks have two parts:

1. Number of bytes that contain skip
2. Skip amount

Skip Size blocks indicate that up to 65,536 contiguous fields aren't directly
encoded in the byte stream. The specific meaning of what the skipped fields
contain is schema and encoding dependent. This is meant to allow efficient
encoding of sparse data (e.g. sparse matrices, delta encoded values).

```
.0              .    1          .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 1|0|     amount    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
size = 1
```

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 1|1|            amount             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
size = 2
```

Example with 1 byte skip size and 16 fields of skip:

```
.0              .    1          .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 1|0|0 0 0 0 1 1 1 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
size = 1
skip = 16
```

### Empty

|              |     |
|--------------|-----|
| Abbreviation | `e` |

Empty blocks indicate that the field is set to the empty value (e.g. an empty
string, integer zero).

```
.0              .
|0 1 2 3 4 5 6 7|
+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 0 1|
+-+-+-+-+-+-+-+-+
```

### Null

|              |     |
|--------------|-----|
| Abbreviation | `n` |

Null blocks indicate that the field is set to the null value (e.g. an optional
string).

```
.0              .
|0 1 2 3 4 5 6 7|
+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 0 0|
+-+-+-+-+-+-+-+-+
```

## Symmetric Blocks

Symmetric blocks can be decoded forwards and reverse. To do so they embed
multi-byte blocks within another type. A naive forward-only decoder can extract
structure and data bits even when symmetric blocks are present, but it cannot
know for certain that the data contains the symmetric closing block in it
without an additional schema (although a heuristic could make a good guess).

* all symmetric blocks have a start and end control block
* (most) end blocks use some of the data bits of the start block
* data and size bits are always read fowards, only the control block processing
  proceeds in reverse
* within byte bits are not reversed, only the order of the control blocks

All single byte sequences are unchanged when symmetric.

### Symmetric Data Size

```
.0              .    1         1.               .5       6      .            7  .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |6 7 8 9 0 1 2 3|4 5 6 7 8 9 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 1|   size    |                1-63 bytes data                |0 1|   size    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

Example Symmetric Data Size block with size 3 (8 + 8 = 16 data bits):

```
.0              .    1          .        2      .            3  .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|4 5 6 7 8 9 0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 1|0 0 0 0 1 0|          2 bytes data         |0 1|0 0 0 0 1 0|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
size = 3
```

Note how the size includes enough bytes to contain the end block. This allows
decoder reading forward to properly find the next field. Decoding in reverse
proceeds in the same fashion: decode last byte, read size of 3, seek backwards
3 bytes to next field.

### Symmetric Data + 1

A Symmetric Data + 1 block (5 + 5 = 10 data bits):

```
.0              .    1          .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 1|  data   |0 0 1|  data   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### Symmetric Data + 2

A Symmetric Data + 2 block (4 + 8 + 4 = 16 data bits):

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 1|         data          |0 0 0 1| data  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### Symmetric Data Size Size

A Symmetric Data Size Size block:

```
.0              .    1         1.               .               .               .               .               .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |               |               |               |               |               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
|0 0 0 0 1|  Z  |         1-8 bytes size        |        1 - ~16 EiB data       |         1-8 bytes size        |0 0 0 0 1|  Z  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
Z: bytes of size
```

Example block with 1 byte of size and 254 bytes of data:

```
                                                                 2   2                   2
                                                                 0   0                   0
.0              .    1         1.        2     2.               .4   5          .        6      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|               |8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 1|0 0 0|1 1 1 1 1 1 1 1|           254 bytes           |1 1 1 1 1 1 1 1|0 0 0 0 1|0 0 0|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
             ^          ^                                               ^                    ^
             |          |                                               |                    |
             |          \                                               /                    |
             |           `---------------* size = 256 *----------------'                     |
             \                                                                               /
              `-----------------------* bytes of size = 1 *---------------------------------'
```

### Symmetric Container

#### Symmetric Bounded

```
.0              .    1         1.               .               .               .               .               .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |               |               |               |               |               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|       size as BSV field       |              BSV              |       size as BSV field       |0 0 0 0 0 1|0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
```

NOTE: The size field in the closing must also be symmetric.

Example block with size 3 (1 data byte):

```
.0              .    1          .        2      .            3  .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|4 5 6 7 8 9 0 1|2 3 4 5 6 7 8 9|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|0|0 0 0 0 0 1 0|1|0 0 0 0 0 0 1|1|0 0 0 0 0 1 0|0 0 0 0 0 1|0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
        ^               ^               ^               ^               ^
        |               |               |               |               |
        |               |               * bsv           |               |
        |               \                               /               |
        |                `--* size = 3 in `d` block *--'                |
        \                                                               /
         `------------------------* `cb` block *-----------------------'
```

Embedded Empty and Null blocks must be expanded to make space for the closing
blocks:

```
.0              .    1          .        2      .            3  .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|4 5 6 7 8 9 0 1|2 3 4 5 6 7 8 9|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|0|0 0 0 0 0 1 0|0 0 0 0 0 0 0 1|1|0 0 0 0 0 1 0|0 0 0 0 0 1|0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

```
.0              .    1          .        2      .            3  .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|4 5 6 7 8 9 0 1|2 3 4 5 6 7 8 9|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|0 1|0|0 0 0 0 0 1 0|0 0 0 0 0 0 0 0|1|0 0 0 0 0 1 0|0 0 0 0 0 1|0 1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

But, we note (as we did in [Container Bounded](#container-bounded)) that this
may be unnecessary if the schema used by the encoder/decoder is sufficiently
expressive.

#### Symmetric Unbounded

```
.0              .    1         1.               .               .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|               |               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
|0 0 0 0 0 1|1 0|              BSV              |0 0 0 0 0 1|0 0|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+++++++++++++++++-+-+-+-+-+-+-+-+
        ^                                               ^
        |                                               |
        * `cu` block                                    * `ce` block
```

Symmetric Unbounded blocks do not need additional closing blocks like the fixed
sized blocks do. The [Container End](#container-end) block is sufficient as the
closing block. This means Symmetric Unbounded blocks are identical to their
non-symmetric form in presentation. The difference is that during reverse
decoding the pair matching logic is inverted.

### Symmetric Skip Size

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 1|1|     amount    |0 0 0 0 0 0 1|1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
size = 2
```

NOTE: Skip Size with size of 1 byte is not possible. There is insufficient
space to put the embedded closing Skip Size block.

Example block with size 2 (skip 16 fields):

```
.0              .    1          .        2      .
|0 1 2 3 4 5 6 7|8 9 0 1 2 3 4 5|6 7 8 9 0 1 2 3|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|0 0 0 0 0 0 1|1|0 0 0 0 1 1 1 1|0 0 0 0 0 0 1|1|
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
size = 2
skip = 16
```
